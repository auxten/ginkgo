package download

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
)

type HttpDownloader struct {
	sync.Once
	client *http.Client
}

func (down *HttpDownloader) DownOffset(uri string, start int64, end int64) (r io.Reader, err error) {
	down.Do(func() {
		down.client = &http.Client{}
	})
	var (
		url    *url.URL
		header http.Header
		resp   *http.Response
	)
	if url, err = url.Parse(uri); err != nil {
		return
	}
	SetByteRange(header, start, end)
	if resp, err = down.client.Do(&http.Request{
		Method: "GET",
		URL:    url,
		Header: header,
		Close:  false,
	}); err != nil {
		return
	}

	r = resp.Body
	return
}

// SetContentRange sets 'Content-Range: bytes startPos-endPos/contentLength'
// header.
func SetContentRange(header http.Header, startPos, endPos, contentLength int64) http.Header {
	if header == nil {
		header = make(http.Header)
	}
	header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", startPos, endPos, contentLength))
	return header
}

// SetByteRange sets 'Range: bytes=startPos-endPos' header.
//
//     * If startPos is negative, then 'bytes=-startPos' value is set.
//     * If endPos is negative, then 'bytes=startPos-' value is set.
func SetByteRange(header http.Header, startPos, endPos int64) http.Header {
	if header == nil {
		header = make(http.Header)
	}
	if startPos < 0 {
		header.Set("Range", fmt.Sprintf("bytes=-%d", -startPos))
	} else if endPos < 0 {
		header.Set("Range", fmt.Sprintf("bytes=%d-", startPos))
	} else {
		header.Set("Range", fmt.Sprintf("bytes=%d-%d", startPos, endPos))
	}
	return header
}
