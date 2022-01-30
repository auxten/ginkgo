package download

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/auxten/ginkgo/seed"
	log "github.com/sirupsen/logrus"
)

type BlockDownloader struct {
	sync.Once
	client *http.Client
	MyHost seed.Host
}

func (down *BlockDownloader) DownBlock(seed *seed.Seed, host string, blockId int64, cnt int64) (r io.ReadCloser, err error) {
	down.Do(func() {
		down.client = &http.Client{}
	})
	var (
		url  *url.URL
		resp *http.Response
	)
	if url, err = url.Parse(fmt.Sprintf("http://%s/api/block?path=%s&id=%d&cnt=%d&host=%s",
		host, seed.Path, blockId, cnt, down.MyHost.String())); err != nil {
		return
	}
	if resp, err = down.client.Do(&http.Request{
		Method: "GET",
		URL:    url,
		Close:  false,
	}); err != nil {
		return
	}

	r = resp.Body
	return
}

func (down *BlockDownloader) BroadcastJoin(host seed.Host, path string, target string) (err error) {
	down.Do(func() {
		down.client = &http.Client{}
	})
	var (
		req  *http.Request
		resp *http.Response
	)

	data, _ := json.Marshal(&seed.HostPath{
		Host: host.String(),
		Path: path,
	})

	req, _ = http.NewRequest("POST",
		fmt.Sprintf("http://%s/api/join", target),
		bytes.NewReader(data),
	)
	req.Header.Set("Content-Type", "application/json")

	if resp, err = down.client.Do(req); err != nil {
		return
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusOK {
		return nil
	} else {
		return fmt.Errorf(resp.Status)
	}
}

func (down *BlockDownloader) GetSeed(host string, uri string, blockSize int64) (sd *seed.Seed, err error) {
	down.Do(func() {
		down.client = &http.Client{}
	})
	var (
		url  *url.URL
		resp *http.Response
		body []byte
	)
	if url, err = url.Parse(fmt.Sprintf("http://%s/api/seed?path=%s&bs=%d&host=%s",
		host, uri, blockSize, down.MyHost.String())); err != nil {
		return
	}
	if resp, err = down.client.Do(&http.Request{
		Method: "GET",
		URL:    url,
		Close:  true,
	}); err != nil {
		return
	}

	defer resp.Body.Close()
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}

	sd = new(seed.Seed)
	if err = json.Unmarshal(body, sd); err != nil {
		return
	}

	for _, h := range sd.Hosts {
		if h == down.MyHost {
			continue
		}
		if er := down.BroadcastJoin(down.MyHost, uri, h.String()); er != nil {
			log.Infof("broadcast join to %s failed: %v", h, er)
		}
	}
	return
}

func (down *BlockDownloader) WriteBlock(sd *seed.Seed, r io.ReadCloser, blockId int64, count int64) (totalWritten int64, err error) {
	var (
		totalSize int64
		bIdx      = blockId
	)

	if blockId+count > int64(len(sd.Blocks)) {
		err = fmt.Errorf("block count cnt %d out of range", count)
		return
	}

	if count <= 0 {
		count = int64(len(sd.Blocks)) - blockId
	}

	for i := blockId; i < blockId+count; i++ {
		if sd.Blocks[i].Done {
			break
		}
		totalSize += sd.Blocks[i].Size
	}

	for fIdx := sd.Blocks[blockId].StartFile; totalWritten < totalSize && fIdx < len(sd.Files); fIdx++ {
		var (
			toWrite        int64
			wrote          int64
			totalRemaining int64
			fileRemaining  int64
			destPath       string
			fd             *os.File
		)

		sFile := sd.Files[fIdx]
		destPath = sFile.LocalPath

		//Add u+w permission
		if sFile.Size == -1 {
			if err = os.MkdirAll(destPath, sFile.Mode.Perm()|0200); err != nil {
				return
			}
		} else if sFile.Size == -2 {
			if err = os.Symlink(sFile.SymPath, destPath); err != nil {
				return
			}
		} else {
			// ensure the first file parent dir exists
			if fIdx == sd.Blocks[blockId].StartFile {
				dir := filepath.Dir(destPath)
				if err = os.MkdirAll(dir, 0755); err != nil {
					return
				}
			}
			if fd, err = os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY, sFile.Mode.Perm()); err != nil {
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

			totalRemaining = totalSize - totalWritten
			if totalRemaining < fileRemaining {
				toWrite = totalRemaining
			} else {
				toWrite = fileRemaining
			}
			if wrote, err = io.CopyN(fd, r, toWrite); err != nil {
				return
			}
			//if err = fd.Sync(); err != nil {
			//	return
			//}
			totalWritten += wrote
			atomic.AddInt64(&sd.TotalWritten, wrote)
			// mark block as written
			for ; totalWritten >= (bIdx-blockId+1)*sd.BlockSize; bIdx++ {
				sd.Blocks[bIdx].Done = true
			}
			if totalWritten == totalSize {
				sd.Blocks[blockId+count-1].Done = true
			}
		}
	}
	return
}
