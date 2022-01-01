package download

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	uri    = "http://nginx.org/download/nginx-1.0.15.tar.gz"
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
		So(hex.EncodeToString(h.Sum(nil)), ShouldResemble, md5sum)
	})

}
func TestHttpDownloader_DownOffset_multi(t *testing.T) {
	Convey("nginx-1.0.15.tar.gz multi ranges", t, func() {
		var (
			n      int
			offset int64
			err    error
			r      io.Reader
		)

		buf := make([]byte, 102400)
		down := HttpDownloader{}
		h := md5.New()
		for err != io.EOF {
			r, err = down.DownOffset(uri, offset, offset+int64(len(buf)))
			So(err, ShouldBeNil)
			for err != io.EOF {
				n, err = r.Read(buf)
				h.Write(buf[:n])
				offset += int64(n)
			}
		}
		So(hex.EncodeToString(h.Sum(nil)), ShouldResemble, md5sum)
	})
}
