package download

import (
	"io"
	"net/http"
	"net/url"
	"sync"
)

type HttpDownloader struct {
	Host string
	sync.Once
	client *http.Client
}

func (down *HttpDownloader) DownOffset(uri string, offset int64) (r io.Reader, err error) {
	down.Do(func() {
		down.client = &http.Client{}
	})
	var url *url.URL
	if url, err = url.Parse(uri); err != nil {
		return
	}
	down.client.Do(&http.Request{
		Method: "GET",
		URL:    url,
		Header: nil,
		Close:  false,
		Host:   down.Host,
	})

}
