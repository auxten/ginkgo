package download

import (
	"io"
)

type DownOffset interface {
	DownOffset(uri string, offset int64) (io.Reader, error)
}
