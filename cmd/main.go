package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/auxten/ginkgo/download"
	"github.com/auxten/ginkgo/fileserv"
	"github.com/auxten/ginkgo/logwrap"
	"github.com/auxten/ginkgo/seed"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
)

var (
	root        string
	addr        string
	srcUri      string
	srcHostPort string
	srcHost     seed.Host
	srcPath     string
	destPath    string
	port        int
	debug       bool
	mode        string
)

func init() {
	flag.StringVar(&root, "root", "./", "Data files server root")
	flag.StringVar(&addr, "addr", "", "Block and http server address, \"0.0.0.0:2120\" for server default, random port on \"0.0.0.0\" for client")
	flag.BoolVar(&debug, "debug", false, "Verbose log for all")
	flag.IntVar(&port, "p", 2120, "Source Host port, like `-P` of scp")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			`Usage of %s:

Server Mode:
  %s -root /path/to/serve

Client Mode:
  %s srcHost:/path/to/src /path/to/dest

By default, client will also serve static files, /path/to/dest is also 
the /path/to/serve, otherwise -root specified.

Options:
`, os.Args[0], os.Args[0], os.Args[0])
		flag.PrintDefaults()
	}
}

func parseFlags() {
	flag.Parse()
	argc := len(flag.Args())
	if argc == 0 { // server mode
		mode = "server"
		if addr == "" {
			addr = "0.0.0.0:2120"
		}
	} else if argc == 2 { // client mode
		mode = "client"
		if addr == "" {
			addr = "0.0.0.0:0"
		}
		srcUri = flag.Arg(0)
		hostPathSlice := strings.Split(srcUri, ":")
		if len(hostPathSlice) != 2 {
			log.Fatalf("Invalid host:path %s", srcUri)
		}
		srcHostPort = fmt.Sprintf("%s:%d", hostPathSlice[0], port)
		var err error
		if srcHost, err = seed.ParseHost(srcHostPort); err != nil {
			log.Fatalf("Invalid srcHost %s: %v", srcHostPort, err)
		}
		srcPath = hostPathSlice[1]
		destPath = flag.Arg(1)
	} else {
		flag.Usage()
		os.Exit(1)
	}
}

func main() {
	parseFlags()

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Logger = logwrap.Logger{Logger: log.StandardLogger()}
	e.Use(logwrap.Hook())

	if debug {
		log.SetLevel(log.DebugLevel)
	}

	e.Use(middleware.Recover())
	log.Debug("starting")
	if mode == "client" {
		var (
			err    error
			lAddr  net.Addr
			myHost seed.Host
			sd     *seed.Seed
		)
		e.POST("/api/join", fileserv.JoinApi())
		e.GET("/api/block", fileserv.BlockApi("./"))
		//wg.Add(1)
		go func() {
			if err = e.Start(addr); err != nil {
				log.Fatal(err)
			}
		}()
		//Fixme
		time.Sleep(time.Second)
		lAddr = e.ListenerAddr()
		hps := strings.Split(lAddr.String(), ":")
		if len(hps) < 2 {
			log.Fatalf("Invalid host:port %s", lAddr.String())
		}
		myHostStr := fmt.Sprintf("%s:%s", GetLocalIP(), hps[len(hps)-1])

		if myHost, err = seed.ParseHost(myHostStr); err != nil {
			log.Fatalf("%s %v", lAddr, err)
		}
		defer e.Close()

		var finished sync.WaitGroup
		finished.Add(1)
		bd := &download.BlockDownloader{MyHost: myHost}
		if sd, err = bd.GetSeed(srcHostPort, srcPath, seed.DefaultBlockSize); err != nil {
			log.Fatalf("GetSeed failed: %v", err)
		}
		if err = sd.Localize(srcPath, destPath); err != nil {
			log.Fatalf("Localize failed: %v", err)
		}
		if err = sd.TouchAll(); err != nil {
			log.Fatalf("TouchAll failed: %v", err)
		}
		blockIndexes := sd.GetBlockIndex(myHost)
		for i := range blockIndexes {
			go func(i int) {
				var (
					startIdx = int64(blockIndexes[i])
					endIdx   = int64(blockIndexes[(i+1)%len(blockIndexes)])
					count    int64
				)
				defer func() {
					// check if all done
					if atomic.LoadInt64(&sd.TotalWritten) == sd.TotalSize {
						finished.Done()
						return
					}
				}()
				log.Debugf("downloading block %d:%d", startIdx, endIdx)
				//PartialDownLoop:
				for {
					hosts := sd.LocateBlock(startIdx, 3)
					// Try srcHost at last
					hosts = append(hosts, srcHost)
					for j, host := range hosts {
						if endIdx == startIdx {
							if len(sd.Blocks) == 1 {
								count = 1
							} else {
								return
							}
						} else if endIdx > startIdx {
							count = endIdx - startIdx
						} else {
							count = -1
						}

						r, er := bd.DownBlock(sd, host.String(), startIdx, count)
						if er == nil {
							defer r.Close()
							wrote, er2 := bd.WriteBlock(sd, r, startIdx, count)
							if wrote > 0 {
								for sd.Blocks[startIdx].Done && startIdx != endIdx {
									startIdx = (startIdx + 1) % int64(len(sd.Blocks))
								}
								if startIdx == endIdx {
									return
								} else {
									break
								}
							} else if er2 != nil {
								if j == len(hosts)-1 {
									log.Fatalf("write block %d failed: %v", startIdx, er2)
								}
							}
						} else {
							if j == len(hosts)-1 {
								log.Fatalf("download block %d end %d count %d failed: %v", startIdx, endIdx, count, er)
							}
						}
					}
				}
			}(i)
		}
		finished.Wait()
		//TODO: broadcast LEAVE
		log.Debugf("%d bytes done!", sd.TotalSize)
		os.Exit(0)
	} else if mode == "server" {
		e.GET("/api/seed", fileserv.SeedApi(root))
		e.POST("/api/join", fileserv.JoinApi())
		e.GET("/api/block", fileserv.BlockApi(root))
		fileserv.ServFiles(e, root)
		e.Logger.Fatal(e.Start(addr))
	}
}

// GetLocalIP returns the non loopback local IP of the host
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
