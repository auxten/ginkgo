package transfer

import (
	log "github.com/auxten/logrus"
	"github.com/auxten/gink-go/src/seed"
	"os"
	"github.com/siddontang/go/num"
	"github.com/Go-zh/tools/container/intsets"
	"github.com/auxten/gink-go/src/util"
)

type BlockReader struct {
	seed             seed.Seed
	startBlockIdx    int64
	fileCursor       int64
	fileFd           *os.File
	fileOffset       int64
	blockCountToRead int64
	byteCountToRead  int64
}

func NewBlockReader(seed seed.Seed, startBlockIdx int64, blockCountToRead int64) (reader *BlockReader) {
	reader = &BlockReader{
		seed:          seed,
		startBlockIdx: startBlockIdx,
	}
	if blockCountToRead < 0 || blockCountToRead > seed.BlockCount {
		// read all block
		reader.blockCountToRead = seed.BlockCount
	}
	reader.byteCountToRead = seed.BlockSize * reader.blockCountToRead
	reader.fileCursor = seed.BlockList[startBlockIdx].FileIndex
	reader.fileOffset = seed.BlockList[startBlockIdx].Offset

	var err error
	path := seed.FileList[reader.fileCursor].Path
	reader.fileFd, err = os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		log.Errorf("OpenFile %s Failed: %s", path, err)
		return nil
	}

	offset, err := reader.fileFd.Seek(reader.fileOffset, 0)
	if offset != reader.fileOffset || err != nil {
		log.Errorf("Seek offset of %s to %d failed: %s", path, reader.fileOffset, err)
		reader.fileFd.Close()
		return nil
	}

	return
}

/*
	+-----------------+-----+-----+-----+-----+-----+-----+-----+-
	|block|block|     |     |     |     |     |     |     |     |
	|  0  |  1  |  2  |  3  |  4  | ... |     |     |     |     |
	+-----+---+-+-----+-----+-----+-----+-----+-----+----++-----+-
	|  file0  | 1|2|3|file4|          Big file5          |  ...
	|         |  | | |     |                             |
	+---------+--+-+-+-----+-----------------------------+--------
*/
func (r *BlockReader) Read(buf []byte) (n int, err error) {
	maxLen := num.MinInt(len(buf), intsets.MaxInt)
	var oneRead int64
	for ; n < maxLen; n += int(oneRead) {
		maxOneRead := num.MinInt64(int64(maxLen - n), r.seed.FileList[r.fileCursor].Size - r.fileOffset)
		r.fileFd.Read(buf[n:])

		// make sure we are opening a file
		for {
			if util.IsFile(r.seed.FileList[r.fileCursor].Mode) {
				break
			}
		}

	}
	return
}

//func main() {
//	M := NewReader("test")
//	stuff, _ := ioutil.ReadAll(M)
//	log.Printf("%s", stuff)
//}
