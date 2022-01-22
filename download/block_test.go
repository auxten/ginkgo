package download

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/auxten/ginkgo/fileserv"
	"github.com/labstack/echo/v4"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBlockDownloader_DownBlock(t *testing.T) {
	e := echo.New()
	e.GET("/api/seed", fileserv.SeedApi("../test"))
	e.GET("/api/block", fileserv.BlockApi("../test"))
	go e.Start("127.0.0.1:0")
	defer e.Close()
	time.Sleep(time.Second)
	addr := e.ListenerAddr()
	bd := &BlockDownloader{}
	sd, _ := bd.GetSeed(addr.String(), "./", 10)

	Convey("block download 0 1", t, func() {
		r, err := bd.DownBlock(sd, addr.String(), 0, 1)
		So(err, ShouldBeNil)
		defer r.Close()
		bytes, err := io.ReadAll(r)
		So(err, ShouldBeNil)
		So(string(bytes), ShouldResemble, "1111111112")
	})

	Convey("block download 0 -1", t, func() {
		r, err := bd.DownBlock(sd, addr.String(), 0, -1)
		So(err, ShouldBeNil)
		defer r.Close()
		bytes, err := io.ReadAll(r)
		So(err, ShouldBeNil)
		So(string(bytes), ShouldResemble, "1111111112222222222222222222223333333333333")
	})

	Convey("block download 1 3", t, func() {
		r, err := bd.DownBlock(sd, addr.String(), 1, 3)
		So(err, ShouldBeNil)
		defer r.Close()
		bytes, err := io.ReadAll(r)
		So(err, ShouldBeNil)
		So(string(bytes), ShouldResemble, "222222222222222222223333333333")
	})
}

//func TestJoin(t *testing.T) {
//	e := echo.New()
//	e.GET("/api/seed", fileserv.SeedApi("../test"))
//	e.GET("/api/join", fileserv.JoinApi())
//	go e.Start("127.0.0.1:0")
//	defer e.Close()
//	time.Sleep(time.Second)
//	addr := e.ListenerAddr()
//	bd := &BlockDownloader{}
//	sd, _ := bd.GetSeed(addr.String(), "./", 10)
//}

func TestBlockDownloader_WriteBlock(t *testing.T) {
	e := echo.New()
	e.GET("/api/seed", fileserv.SeedApi("../test"))
	e.GET("/api/block", fileserv.BlockApi("../test"))
	go e.Start("127.0.0.1:0")
	defer e.Close()
	time.Sleep(time.Second)
	addr := e.ListenerAddr()
	bd := &BlockDownloader{}

	Convey("block write 0 1", t, func() {
		dir, err := ioutil.TempDir("", "blockWrite")
		So(err, ShouldBeNil)
		defer os.RemoveAll(dir) // clean up
		sd, err := bd.GetSeed(addr.String(), "./", 10)
		So(err, ShouldBeNil)
		r, err := bd.DownBlock(sd, addr.String(), 0, 1)
		So(err, ShouldBeNil)
		defer r.Close()
		err = bd.WriteBlock("./", dir, sd, r, 0, 1)
		So(err, ShouldBeNil)
		bytes, err := os.ReadFile(filepath.Join(dir, "dir1/file11"))
		So(err, ShouldBeNil)
		So(string(bytes), ShouldResemble, "111111111")
		bytes, err = os.ReadFile(filepath.Join(dir, "dir1/file12"))
		So(err, ShouldBeNil)
		So(string(bytes), ShouldResemble, "2")
		So(sd.Blocks[0].Done, ShouldBeTrue)
		So(sd.Blocks[1].Done, ShouldBeFalse)
	})

	Convey("block write 0 -1", t, func() {
		dir, err := ioutil.TempDir("", "blockWrite")
		So(err, ShouldBeNil)
		//defer os.RemoveAll(dir) // clean up
		sd, err := bd.GetSeed(addr.String(), "./", 10)
		So(err, ShouldBeNil)
		r, err := bd.DownBlock(sd, addr.String(), 0, -1)
		So(err, ShouldBeNil)
		defer r.Close()
		err = bd.WriteBlock("./", dir, sd, r, 0, -1)
		So(err, ShouldBeNil)
		bytes, err := os.ReadFile(filepath.Join(dir, "dir1/file11"))
		So(err, ShouldBeNil)
		So(string(bytes), ShouldResemble, "111111111")
		bytes, err = os.ReadFile(filepath.Join(dir, "dir1/file12"))
		So(err, ShouldBeNil)
		So(string(bytes), ShouldResemble, "222222222222222222222")
		bytes, err = os.ReadFile(filepath.Join(dir, "dir1/file13"))
		So(err, ShouldBeNil)
		So(string(bytes), ShouldResemble, "3333333333333")
		for _, b := range sd.Blocks {
			So(b.Done, ShouldBeTrue)
		}
	})

	Convey("block write 1 3", t, func() {
		dir, err := ioutil.TempDir("", "blockWrite")
		So(err, ShouldBeNil)
		defer os.RemoveAll(dir) // clean up
		sd, err := bd.GetSeed(addr.String(), "./", 10)
		So(err, ShouldBeNil)
		r, err := bd.DownBlock(sd, addr.String(), 1, 3)
		So(err, ShouldBeNil)
		defer r.Close()
		err = bd.WriteBlock("./", dir, sd, r, 1, 3)
		So(err, ShouldBeNil)
		bytes, err := os.ReadFile(filepath.Join(dir, "dir1/file11"))
		So(err.Error(), ShouldContainSubstring, "no such file or directory")
		bytes, err = os.ReadFile(filepath.Join(dir, "dir1/file12"))
		So(err, ShouldBeNil)
		So(string(bytes), ShouldResemble, "\x0022222222222222222222")
		bytes, err = os.ReadFile(filepath.Join(dir, "dir1/file13"))
		So(err, ShouldBeNil)
		So(string(bytes), ShouldResemble, "3333333333")
		for i, b := range sd.Blocks {
			if i >= 1 && i <= 3 {
				So(b.Done, ShouldBeTrue)
			} else {
				So(b.Done, ShouldBeFalse)
			}
		}
	})

}
