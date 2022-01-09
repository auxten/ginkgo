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
	log "github.com/auxten/logrus"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

type writeFlusher interface {
	http.Flusher
	http.ResponseWriter
}

var (
	sds struct {
		sync.Map // map[path]*Seed
	}
)

func ServFiles(e *echo.Echo, root string) {
	e.Static(root, "f")
	return
}

func SeedApi(root string) func(echo.Context) error {
	return func(c echo.Context) (err error) {
		var (
			bs   int64 = -1
			sd   *seed.Seed
			path string
		)
		path = c.QueryParam("path")
		if len(path) == 0 {
			return c.String(http.StatusNotFound, "need path query param")
		}
		if val, ok := sds.Load(path); ok {
			sd = val.(*seed.Seed)
			log.Debugf("cached seed: %s", path)
		} else {
			if blockSize := c.QueryParam("bs"); len(blockSize) != 0 {
				if bs, err = strconv.ParseInt(blockSize, 10, 64); err != nil {
					return c.String(http.StatusBadRequest,
						fmt.Sprintf("invalid block Size bs: %v", blockSize))
				}
			}
			jailedPath := filepath.Join(root, path)
			log.Debugf("making seed for path: %s", jailedPath)
			if sd, err = seed.MakeSeed(jailedPath, bs); err != nil {
				return c.String(http.StatusInternalServerError, err.Error())
			}
			sds.Store(path, sd)
		}
		return c.JSON(http.StatusOK, sd)
	}
}

func BlockApi(root string) func(echo.Context) error {
	return func(c echo.Context) (err error) {
		var (
			path    string
			id      string
			sd      *seed.Seed
			blockId int64
			count   int64 // 0 means infinite till last block
		)
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

func sendBlock(blockId int64, count int64, sd *seed.Seed, root string, respWriter writeFlusher) (err error) {
	var (
		totalSize int64
		totalSent int64
	)
	if blockId+count > int64(len(sd.Blocks)) {
		return errors.Errorf("block count cnt %d out of range", count)
	}
	if count <= 0 {
		count = int64(len(sd.Blocks)) - blockId
	}

	for i := blockId; i < blockId+count; i++ {
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
		jailedPath := filepath.Join(root, sFile.Path)

		if fd, err = os.Open(jailedPath); err != nil {
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
