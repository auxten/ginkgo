package fileserv

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/auxten/ginkgo/seed"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

type writeFlusher interface {
	http.Flusher
	http.ResponseWriter
}

const VNodeCount = uint8(3)

var (
	sds struct {
		sync.Map // map[path]*Seed
	}
)

func SeedApi(root string) func(echo.Context) error {
	if er := os.Chdir(root); er != nil {
		log.Fatal(er)
	}

	return func(c echo.Context) (err error) {
		var (
			bs      int64 = -1
			sd      *seed.Seed
			path    string
			hostStr string
			host    seed.Host
		)
		hostStr = c.QueryParam("host")
		if len(hostStr) == 0 {
			return c.String(http.StatusBadRequest, "need host query param")
		}
		if host, err = seed.ParseHost(hostStr); err != nil {
			return c.String(http.StatusBadRequest, fmt.Sprintf("bad host query param %s", err))
		}

		path = c.QueryParam("path")
		if len(path) == 0 {
			return c.String(http.StatusBadRequest, "need path query param")
		}
		cleanPath := filepath.Clean(path)
		if val, ok := sds.Load(cleanPath); ok {
			sd = val.(*seed.Seed)
			sd.Hosts = sd.GetAllHosts()
			log.Debugf("cached seed: %s", path)
		} else {
			if blockSize := c.QueryParam("bs"); len(blockSize) != 0 {
				if bs, err = strconv.ParseInt(blockSize, 10, 64); err != nil {
					return c.String(http.StatusBadRequest,
						fmt.Sprintf("invalid block Size bs: %v", blockSize))
				}
			}
			log.Debugf("making seed for path: %s", cleanPath)
			if sd, err = seed.MakeSeed(cleanPath, bs); err != nil {
				return c.String(http.StatusInternalServerError, err.Error())
			}
			sd.Path = cleanPath
			sd.VNodeCount = VNodeCount

			sds.Store(cleanPath, sd)
		}
		sd.Add(host)
		return c.JSON(http.StatusOK, sd)
	}
}

func BlockApi(root string) func(echo.Context) error {
	if er := os.Chdir(root); er != nil {
		log.Fatal(er)
	}

	return func(c echo.Context) (err error) {
		var (
			path    string
			id      string
			sd      *seed.Seed
			blockId int64
			count   int64 // 0 means infinite till last block
			hostStr string
			host    seed.Host
		)
		hostStr = c.QueryParam("host")
		if len(hostStr) == 0 {
			return c.String(http.StatusBadRequest, "need host query param")
		}
		if host, err = seed.ParseHost(hostStr); err != nil {
			return c.String(http.StatusBadRequest, fmt.Sprintf("bad host query param %s", err))
		}

		path = c.QueryParam("path")
		if len(path) == 0 {
			return c.String(http.StatusNotFound, "need path query param")
		}
		if val, ok := sds.Load(path); !ok {
			return c.String(http.StatusNotFound,
				fmt.Sprintf("seed for path %s not found", path))
		} else {
			sd = val.(*seed.Seed)
		}
		sd.Add(host)

		id = c.QueryParam("id")
		if len(id) == 0 {
			return c.String(http.StatusBadRequest, "need blockId param")
		}
		if blockId, err = strconv.ParseInt(id, 10, 64); err != nil {
			return c.String(http.StatusBadRequest,
				fmt.Sprintf("invalid block ID blockId %v", id))
		}
		if blockId > int64(len(sd.Blocks)-1) {
			return c.String(http.StatusBadRequest,
				fmt.Sprintf("block ID %d out of range", blockId))
		}

		if cnt := c.QueryParam("cnt"); len(cnt) != 0 {
			if count, err = strconv.ParseInt(cnt, 10, 64); err != nil {
				return c.String(http.StatusBadRequest,
					fmt.Sprintf("invalid block count cnt: %v", cnt))
			}
		}
		if blockId+count > int64(len(sd.Blocks)) {
			return c.String(http.StatusBadRequest,
				fmt.Sprintf("block count cnt %d out of range", count))
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEOctetStream)
		c.Response().WriteHeader(http.StatusOK)
		respWriter := c.Response()

		if err = sendBlock(blockId, count, sd, root, respWriter); err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return
	}
}

//JoinApi puts the host into the corresponding seed of path, and return all known hosts in the seed.
func JoinApi() func(echo.Context) error {
	return func(c echo.Context) (err error) {
		var (
			sd       *seed.Seed
			host     seed.Host
			hostPath = new(seed.HostPath)
		)
		if err = c.Bind(hostPath); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}

		if len(hostPath.Host) == 0 {
			return c.String(http.StatusBadRequest, "need host query param")
		}
		if host, err = seed.ParseHost(hostPath.Host); err != nil {
			return c.String(http.StatusBadRequest, fmt.Sprintf("bad host query param %s", err))
		}
		if len(hostPath.Path) == 0 {
			return c.String(http.StatusBadRequest, "need path query param")
		}

		cleanPath := filepath.Clean(hostPath.Path)
		if val, ok := sds.Load(cleanPath); !ok {
			return c.String(http.StatusNotFound, fmt.Sprintf("seed for path: %s not found", cleanPath))
		} else {
			sd = val.(*seed.Seed)
			sd.Add(host)
			return c.JSON(http.StatusOK, sd.GetAllHosts())
		}
	}
}

func sendBlock(blockId int64, count int64, sd *seed.Seed, root string, respWriter writeFlusher) (err error) {
	var (
		totalSize int64
		totalSent int64
	)
	if blockId+count > int64(len(sd.Blocks)) {
		return fmt.Errorf("block count cnt %d out of range", count)
	}
	if count <= 0 {
		count = int64(len(sd.Blocks)) - blockId
	}

	for i := blockId; i < blockId+count; i++ {
		if !sd.Blocks[i].Done {
			break
		}
		totalSize += sd.Blocks[i].Size
	}
	for fIdx := sd.Blocks[blockId].StartFile; totalSent < totalSize && fIdx < len(sd.Files); fIdx++ {
		var (
			toSend         int64
			sent           int64
			totalRemaining int64
			fileRemaining  int64
			fd             *os.File
		)

		sFile := sd.Files[fIdx]

		if fd, err = os.Open(sFile.LocalPath); err != nil {
			return
		}
		defer fd.Close()
		// the first file should seek to the block offset
		if fIdx == sd.Blocks[blockId].StartFile {
			fileRemaining = sFile.Size - sd.Blocks[blockId].StartOffset
			if _, err = fd.Seek(sd.Blocks[blockId].StartOffset, io.SeekStart); err != nil {
				return
			}
		} else {
			fileRemaining = sFile.Size
		}

		totalRemaining = totalSize - totalSent
		if totalRemaining < fileRemaining {
			toSend = totalRemaining
		} else {
			toSend = fileRemaining
		}
		if sent, err = io.CopyN(respWriter, fd, toSend); err != nil {
			return
		}
		respWriter.Flush()
		totalSent += sent
	}
	return
}
