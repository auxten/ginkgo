package seed

import (
	"crypto/sha256"
	"fmt"
	"hash/crc32"
	"hash/fnv"
	"net/netip"
	"os"
	"sync"
	"time"

	"github.com/auxten/ginkgo/srcdest"
)

var (
	DefaultBlockSize int64 = 4 * 1024 * 1024
)

type Seed struct {
	sync.RWMutex
	Path       string   `json:"path"`
	FileCount  int      `json:"fileCount"`
	Files      []*File  `json:"files"`
	Blocks     []*Block `json:"blocks"`
	BlockSize  int64    `json:"blockSize"`
	VNodeCount uint8    `json:"vnodeCount"`
	// Hosts are only updated before Marshal
	Hosts []Host `json:"hosts"`
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
	SymPath   string `json:"symPath"`
	Path      string `json:"path"`
	LocalPath string `json:"-"`
	CheckSum  []byte `json:"checkSum"`
}

type Host struct {
	IP   [4]byte `json:"ip"` // IPv4 only
	Port uint16  `json:"port"`
}

type HostPath struct {
	Host string `json:"host" form:"host" query:"host"`
	Path string `json:"path" form:"path" query:"path"`
}

type Block struct {
	Size        int64              `json:"size"`
	StartFile   int                `json:"startFile"`
	StartOffset int64              `json:"startOffset"`
	Done        bool               `json:"-"`
	CheckSum    []byte             `json:"checkSum"`
	Hosts       map[Host]time.Time `json:"-"`
}

func (h Host) String() string {
	return fmt.Sprintf("%d.%d.%d.%d:%d", h.IP[0], h.IP[1], h.IP[2], h.IP[3], h.Port)
}

// ParseHost parses "IPv4:Port"
func ParseHost(hStr string) (Host, error) {
	if ipPort, err := netip.ParseAddrPort(hStr); err != nil {
		return Host{}, err
	} else {
		if !ipPort.Addr().Is4() {
			return Host{}, fmt.Errorf("only IPv4 addresses")
		}
		ipBytes, _ := ipPort.Addr().MarshalBinary()
		return Host{
			IP:   [4]byte{ipBytes[0], ipBytes[1], ipBytes[2], ipBytes[3]},
			Port: ipPort.Port(),
		}, nil
	}
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

func (sd *Seed) Localize(cmdSrcPath string, cmdDestPath string) (err error) {
	var (
		srcType  srcdest.PathType
		destType srcdest.PathType
		fInfo    os.FileInfo
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

	for i := range sd.Files {
		sd.Files[i].LocalPath, err = srcdest.NormalizeDestPath(cmdSrcPath, cmdDestPath, srcType, destType, sd.Files[i].Path)
		if err != nil {
			return
		}
	}

	return
}
