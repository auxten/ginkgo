package download

import (
	"io"

	"github.com/auxten/ginkgo/seed"
)

type DownHttp interface {
	DownOffset(uri string, start int64, end int64) (io.ReadCloser, error)
}

type DownBlock interface {
	GetSeed(host string, uri string, blockSize int64) (*seed.Seed, error)
	DownBlock(seed *seed.Seed, host string, blockId int64, cnt int64) (io.ReadCloser, error)
}
