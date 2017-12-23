package transfer

import (
	"bufio"
	log "github.com/auxten/logrus"
	"github.com/auxten/gink-go/src/seed"
	"os"
)

type BlockReader interface {
	// 1. 通过RPC告知对端要从blockIndex开始收blockCount个块
	// 2. 获取一个bufio.Reader
	Read(blockIndex int64, blockCount int64) (reader bufio.Reader, err error)
}

type BlockWriter interface {
	// 1. 通过blockindex确定应该在哪个文件的位置开始写
	Write(blockIndex int64, blockCount int64) (writer bufio.Writer, err error)
}

type BlockFileIO interface {
	BlockReader
	BlockWriter
}

type BlockSocketIO interface {
	BlockReader
	BlockWriter
}

type BlockServer struct {
	seed seed.Seed
	BlockFileIO
	BlockSocketIO
}

func (b *BlockServer) Read(blockIndex int64, blockCount int64) (reader bufio.Reader, err error) {

	return
}

func (b *BlockServer) Write(blockIndex int64, blockCount int64) (reader bufio.Writer, err error) {
	return
}

/*
	scp -r ./src_dir/src xx.com:./dst_dir/dst
	if src is Dir and dst is dir
		./dst_dir/dst/src
	if src is Dir and dst not exist
		./dst_dir/dst
	if src is Dir and dst is file
		fail
	if src is file and dst is file
		overwrite
	if src is file and dst is dir
		./dst_dir/dst/src

	scp -r ./src_dir/src xx.com:./dst_dir/dst/
	if src is Dir and dst is dir
		./dst_dir/dst/src
	if src is Dir and dst is file
		fail
	if src is file and dst is file
		fail
	if src is file and dst is dir
		./dst_dir/dst/src
 */

func (b *BlockServer) DownloadBlock(startBlockIndex int64, blockCount int64) (count int64, err error) {
	socketReader, err := b.BlockSocketIO.Read(startBlockIndex, blockCount)
	if err != nil {
		log.Errorf("get block from socket error, idx:%d, count:%d", startBlockIndex, blockCount)
	}
	wblock := b.seed.BlockList[startBlockIndex]
	for remainSize := b.seed.BlockSize * blockCount; remainSize > 0; {
		fileIndex := wblock.StartFileIndex
		fileOffset := wblock.StartOffset
		wfile := b.seed.FileList[fileIndex]
		// todo 创建本文件所有需要的文件夹，defer-close

		/* todo
			打开文件，定位到块起始的位置，写入数据，不要超出原有文件大小
		 	写完一个文件就进行下一次循环，主要是fileIndex
			注意对remainSize进行减小
		*/
	}

	return
}
