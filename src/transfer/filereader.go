package transfer

import (
	"os"
)

type TransData interface {
	Read(r *os.File) error
	Write(w *os.File) error
}

type TransFile struct {
	buf []byte
}

func (f TransFile) Read(r *os.File) (error) {
	return nil
}

func (f TransFile) Write(w *os.File) (error) {

	return nil
}
