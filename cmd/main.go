package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/auxten/ginkgo/fileserv"
	"github.com/auxten/ginkgo/logwrap"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
)

var (
	root  string
	addr  string
	debug bool
	mode  string
)

func init() {
	flag.StringVar(&root, "root", "./", "Data files server root")
	flag.StringVar(&addr, "addr", ":2120", "Data files output dir and also http static root")
	flag.BoolVar(&debug, "debug", false, "Verbose log for all")

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
	} else if argc == 2 { // client mode
		mode = "client"
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

	e.GET("/api/seed", fileserv.SeedApi(root))
	e.POST("/api/join", fileserv.JoinApi())
	e.GET("/api/block", fileserv.BlockApi(root))

	fileserv.ServFiles(e, root)

	e.Logger.Fatal(e.Start(addr))
}
