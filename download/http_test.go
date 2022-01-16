package download

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	uri = "http://nginx.org/download/nginx-1.0.15.tar.gz"
	//uri    = "http://localhost:9099/nginx-1.0.15.tar.gz"
	//uri    = "http://localhost:1323/nginx"
	md5sum = "17da4802209b83d9bebb0f0edd975dfc"
)

func TestHttpDownloader_DownOffset(t *testing.T) {
	Convey("nginx-1.0.15.tar.gz", t, func() {
		var (
			n int
		)

		buf := make([]byte, 102400)
		down := HttpDownloader{}
		r, err := down.DownOffset(uri, 0, -1)
		So(err, ShouldBeNil)
		h := md5.New()
		for err != io.EOF {
			n, err = r.Read(buf)
			h.Write(buf[:n])
		}
		_ = r.Close()
		So(hex.EncodeToString(h.Sum(nil)), ShouldResemble, md5sum)
	})

	Convey("nginx-1.0.15.tar.gz", t, func() {
		var (
			n int
		)

		buf := make([]byte, 102400)
		down := HttpDownloader{}
		r, err := down.DownOffset(uri, 10, -1)
		So(err, ShouldBeNil)
		h := md5.New()
		for err != io.EOF {
			n, err = r.Read(buf)
			h.Write(buf[:n])
		}
		_ = r.Close()
		So(hex.EncodeToString(h.Sum(nil)), ShouldResemble, "23c12b79ca4dee716f43b9fef10b296d")
	})

	Convey("nginx-1.0.15.tar.gz", t, func() {
		var (
			n int
		)

		buf := make([]byte, 102400)
		down := HttpDownloader{}
		r, err := down.DownOffset(uri, 10, 11) //echo "ECBD" | xxd -r -p| md5
		So(err, ShouldBeNil)
		h := md5.New()
		for err != io.EOF {
			n, err = r.Read(buf)
			h.Write(buf[:n])
		}
		_ = r.Close()
		So(hex.EncodeToString(h.Sum(nil)), ShouldResemble, "08581bfce36b88261f1158c2b01efa82")
	})

}

func TestHttpDownloader_DownOffset_multi(t *testing.T) {
	Convey("nginx-1.0.15.tar.gz multi ranges", t, func() {
		var (
			n      int
			acc    int
			offset int64
			err    error
			r      io.Reader
		)

		buf := make([]byte, 102400)
		down := HttpDownloader{}
		h := md5.New()
		for err != io.EOF {
			acc = 0
			r, err = down.DownOffset(uri, offset, offset+int64(len(buf)))
			So(err, ShouldBeNil)
			for err != io.EOF {
				if n, err = r.Read(buf[acc:]); err != nil {
					return
				}
				acc += n
				if acc == len(buf) {
					break
				}
			}
			h.Write(buf[:acc])
			offset += int64(acc)
		}
		So(hex.EncodeToString(h.Sum(nil)), ShouldResemble, md5sum)
	})
}
