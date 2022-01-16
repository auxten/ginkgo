package download

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"

	"github.com/auxten/ginkgo/seed"
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
		Close:  false,
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

//func (down *BlockDownloader) WriteBlock(r io.ReadCloser, dest string, seed *seed.Seed, blockId int64, cnt int64) (err error) {
//
//}
