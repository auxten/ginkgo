package seed

import (
	"crypto/sha256"
	"fmt"
	"hash/crc32"
	"hash/fnv"
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
	Size        int64         `json:"size"`
	StartFile   int           `json:"startFile"`
	StartOffset int64         `json:"startOffset"`
	Done        bool          `json:"-"`
	CheckSum    []byte        `json:"checkSum"`
	Hosts       map[Host]bool `json:"-"`
}

func (h Host) String() string {
	return fmt.Sprintf("%d.%d.%d.%d:%d", h.IP[0], h.IP[1], h.IP[2], h.IP[3], h.Port)
}

// Hash uses HashCrc which has better uniformity
func (h Host) Hash(vIndex byte) uint32 {
	return h.HashCrc(vIndex)
}

func (h Host) HashFnv(vIndex byte) uint32 {
	hash := fnv.New32a()
	_, _ = hash.Write(h.IP[:])
	_, _ = hash.Write([]byte{byte(h.Port / 256), byte(h.Port % 256), vIndex})
	return hash.Sum32()
}

func (h Host) HashCrc(vIndex byte) uint32 {
	hash := crc32.New(crc32.MakeTable(crc32.Castagnoli))
	_, _ = hash.Write(h.IP[:])
	_, _ = hash.Write([]byte{byte(h.Port / 256), byte(h.Port % 256), vIndex})
	return hash.Sum32()
}

func (h Host) HashSha(vIndex byte) uint32 {
	hash := sha256.New()
	_, _ = hash.Write(h.IP[:])
	_, _ = hash.Write([]byte{byte(h.Port / 256), byte(h.Port % 256), vIndex})
	hout := fnv.New32a()
	hout.Write(hash.Sum(nil))
	return hout.Sum32()
}
