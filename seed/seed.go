package seed

import (
	"os"
	"time"
)

var (
	DefaultBlockSize int64 = 4 * 1024 * 1024
)

type Seed struct {
	Path      string   `json:"path"`
	FileCount int      `json:"fileCount"`
	Files     []*File  `json:"files"`
	Blocks    []*Block `json:"blocks"`
	BlockSize int64    `json:"blockSize"`
	//InitFileIdx   int      `json:"initFileIdx"`
	//InitBlockIdx  int      `json:"initBlockIdx"`
	TotalSize int64 `json:"totalSize"`
	//TmpSize       int64    `json:"tmpSize"`
	//LastInitBlock int64    `json:"lastInitBlock"`
}

type File struct {
	mtime time.Time
	Mode  os.FileMode `json:"mode"`
	//Size of file, -1 for dir, -2 for symbol link
	Size int64 `json:"size"`
	//SymPath if symbol link target path
	SymPath  string `json:"symPath"`
	Path     string `json:"path"`
	CheckSum []byte `json:"checkSum"`
}

type Host struct {
	IP   [4]byte `json:"ip"` // IPv4 only
	Port uint16  `json:"port"`
}

type Block struct {
	Size        int64  `json:"size"`
	StartFile   int    `json:"startFile"`
	StartOffset int64  `json:"startOffset"`
	Done        bool   `json:"-"`
	CheckSum    []byte `json:"checkSum"`
	hosts       map[Host]bool
}
