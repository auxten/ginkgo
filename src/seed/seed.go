package seed

import (
	"os"
	"path/filepath"
	log "github.com/auxten/logrus"
	//"github.com/dixonwille/skywalker"
)

type FileType int

type FileInfo struct {
	Path string
	Size int64
	Mode os.FileMode // This contains file type
}

type BlockInfo struct {
	StartFileIndex int64
	StartOffset    int64
}

type Seed struct {
	BlockSize  int64
	BlockCount int64
	FileList   []FileInfo
	BlockList  []BlockInfo
}

func NewSeed(blockSize int64) Seed {
	return Seed{
		BlockSize:  blockSize,
		BlockCount: -1,
		FileList:   []FileInfo{},
		BlockList:  []BlockInfo{},
	}
}

func (s *Seed) MakeSeed(rootPath string) (err error) {
	var totalFileSize int64
	filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		log.Debug(info.Mode()&os.ModeType == 0, path)
		s.FileList = append(s.FileList, FileInfo{
			Path: path,
			Size: info.Size(),
			Mode: info.Mode(),
		})
		if info.Mode()&os.ModeType == 0 {
			totalFileSize += info.Size()
		}
		return nil
	})

	log.Debug(s.FileList)
	log.Debug("totalFileSize ", totalFileSize)

	//var (
	//	totalBlockSize int64
	//	lastBlockIdx   int64 = -1
	//)
	/*
		+-----------------+-----+-----+-----+-----+-----+-----+-----+-
		|block|block|     |     |     |     |     |     |     |     |
		|  0  |  1  |  2  |  3  |  4  | ... |     |     |     |     |
		+-----+---+-+-----+-----+-----+-----+-----+-----+----++-----+-
		|  file0  | 1|2|3|file4|          Big file5          |  ...
		|         |  | | |     |                             |
		+---------+--+-+-+-----+-----------------------------+--------
	*/
	//for blockIdx, block := range s.BlockList {
	//	if totalFileSize > totalBlockSize {
	//		totalBlockSize += s.BlockSize
	//		for fileIdx, file := range s.FileList {
	//
	//		}
	//	}
	//}
	return
}
