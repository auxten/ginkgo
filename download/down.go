package download

import (
	"io"
)

type DownOffset interface {
	DownOffset(uri string, start int64, end int64) (io.Reader, error)
}
