package download

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	"github.com/auxten/ginkgo/seed"
	"github.com/auxten/ginkgo/srcdest"
)

type BlockDownloader struct {
	sync.Once
	client *http.Client
}

func (down *BlockDownloader) DownBlock(seed *seed.Seed, host string, blockId int64, cnt int64) (r io.ReadCloser, err error) {
	down.Do(func() {
		down.client = &http.Client{}
	})
	var (
		url  *url.URL
		resp *http.Response
	)
	if url, err = url.Parse(fmt.Sprintf("http://%s/api/block?path=%s&id=%d&cnt=%d",
		host, seed.Path, blockId, cnt)); err != nil {
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

func (down *BlockDownloader) GetSeed(host string, uri string, blockSize int64) (s *seed.Seed, err error) {
	down.Do(func() {
		down.client = &http.Client{}
	})
	var (
		url  *url.URL
		resp *http.Response
		body []byte
	)
	if url, err = url.Parse(fmt.Sprintf("http://%s/api/seed?path=%s&bs=%d",
		host, uri, blockSize)); err != nil {
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

	s = new(seed.Seed)
	if err = json.Unmarshal(body, s); err != nil {
		return
	}

	return
}

func (down *BlockDownloader) WriteBlock(
	cmdSrcPath string, cmdDestPath string, sd *seed.Seed, r io.ReadCloser, blockId int64, count int64,
) (err error) {

	var (
		totalSize    int64
		totalWritten int64
		srcType      srcdest.PathType
		destType     srcdest.PathType
		fInfo        os.FileInfo
		bIdx         = blockId
	)
	if sd.Files[0].Size >= 0 {
		srcType = srcdest.FileType
	} else if sd.Files[0].Size == -1 {
		srcType = srcdest.DirType
	} else {
		return fmt.Errorf("src root path type %d is not supported", sd.Files[0].Size)
	}

	if fInfo, err = os.Stat(cmdDestPath); err != nil {
		if err == os.ErrNotExist {
			destType = srcdest.NotExist
		} else {
			return
		}
	} else if fInfo.IsDir() {
		destType = srcdest.DirType
	} else if fInfo.Mode().IsRegular() {
		destType = srcdest.FileType
	} else {
		return fmt.Errorf("dest path type %s is not supported", fInfo.Mode().Type().String())
	}

	if blockId+count > int64(len(sd.Blocks)) {
		return fmt.Errorf("block count cnt %d out of range", count)
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
		destPath, err = srcdest.NormalizeDestPath(cmdSrcPath, cmdDestPath, srcType, destType, sFile.Path)
		if err != nil {
			return
		}

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
			if err = fd.Sync(); err != nil {
				return
			}
			totalWritten += wrote
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
