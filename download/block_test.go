package download

import (
	"io"
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
	time.Sleep(time.Second)
	addr := e.ListenerAddr()
	bd := &BlockDownloader{}
	seed, _ := bd.GetSeed(addr.String(), "./", 10)

	Convey("block download", t, func() {
		r, err := bd.DownBlock(seed, addr.String(), 0, 1)
		So(err, ShouldBeNil)
		defer r.Close()
		bytes, err := io.ReadAll(r)
		So(err, ShouldBeNil)
		So(string(bytes), ShouldResemble, "1111111112")
	})

	Convey("block download", t, func() {
		r, err := bd.DownBlock(seed, addr.String(), 0, -1)
		So(err, ShouldBeNil)
		defer r.Close()
		bytes, err := io.ReadAll(r)
		So(err, ShouldBeNil)
		So(string(bytes), ShouldResemble, "1111111112222222222222222222223333333333333")
	})

	Convey("block download", t, func() {
		r, err := bd.DownBlock(seed, addr.String(), 1, 3)
		So(err, ShouldBeNil)
		defer r.Close()
		bytes, err := io.ReadAll(r)
		So(err, ShouldBeNil)
		So(string(bytes), ShouldResemble, "222222222222222222223333333333")
	})
}
