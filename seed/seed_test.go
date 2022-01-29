package seed

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	log "github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

func TestHost_Hash(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	const (
		size      = 256
		vCount    = 5
		threshold = 20
	)

	var testData = []struct {
		host Host
	}{
		{Host{[4]byte{0, 0, 0, 0}, 8080}},
		{Host{[4]byte{127, 0, 0, 1}, 8080}},
		{Host{[4]byte{192, 168, 0, 1}, 8080}},
		{Host{[4]byte{192, 168, 0, 2}, 8080}},
		{Host{[4]byte{192, 168, 0, 3}, 8080}},
		{Host{[4]byte{10, 10, 0, 3}, 8080}},
		{Host{[4]byte{10, 20, 0, 3}, 8080}},
		{Host{[4]byte{10, 20, 0, 4}, 8080}},
		{Host{[4]byte{10, 20, 0, 4}, 8081}},
		{Host{[4]byte{10, 20, 0, 4}, 8082}},
		{Host{[4]byte{10, 20, 1, 4}, 8082}},
	}
	t.Run("fnv hash collision test", func(t *testing.T) {
		var pool = make(map[uint64]byte)
		var counter int
		for _, d := range testData {
			for j := byte(0); j < vCount; j++ {
				h := uint64(d.host.HashFnv(j) % size)
				//log.Debugf("%d", h)
				if _, collision := pool[h]; collision {
					log.Debugf("%s %d found %d", d.host, j, h)
					counter++
				} else {
					pool[h] = 1
				}
			}
		}
		if counter > threshold {
			t.Errorf("collision count %d", counter)
		}
		log.Debugf("fnv collision count %d", counter)
	})
	t.Run("crc32 hash collision test", func(t *testing.T) {
		var pool = make(map[uint64]byte)
		var counter int
		for _, d := range testData {
			for j := byte(0); j < vCount; j++ {
				h := uint64(d.host.HashCrc(j) % size)
				//log.Debugf("%d", h)
				if _, collision := pool[h]; collision {
					log.Debugf("%s %d found %d", d.host, j, h)
					counter++
				} else {
					pool[h] = 1
				}
			}
		}
		if counter > threshold {
			t.Errorf("collision count %d", counter)
		}
		log.Debugf("crc32 collision count %d", counter)
	})
	t.Run("sha256 hash collision test", func(t *testing.T) {
		var pool = make(map[uint64]byte)
		var counter int
		for _, d := range testData {
			for j := byte(0); j < vCount; j++ {
				h := uint64(d.host.HashSha(j) % size)
				//log.Debugf("%d", h)
				if _, collision := pool[h]; collision {
					log.Debugf("%s %d found %d", d.host, j, h)
					counter++
				} else {
					pool[h] = 1
				}
			}
		}
		if counter > threshold {
			t.Errorf("collision count %d", counter)
		}
		log.Debugf("sha256 collision count %d", counter)
	})
}

func TestHost_String(t *testing.T) {
	t.Run("hash string", func(t *testing.T) {
		h := Host{
			IP:   [4]byte{10, 20, 30, 110},
			Port: 8081,
		}
		if h.String() != "10.20.30.110:8081" {
			t.Fatal()
		}
	})
}

func TestSeed_TouchAll(t *testing.T) {
	Convey("ensure touch all", t, func() {
		dir, err := ioutil.TempDir("", "touchAll")
		So(err, ShouldBeNil)
		defer os.RemoveAll(dir) // clean up
		err = os.Chdir(dir)
		So(err, ShouldBeNil)
		sd := Seed{
			Files: []*File{
				{LocalPath: "./dir", Size: -1},
				{LocalPath: "./dir/dir1", Size: -1},
				{LocalPath: "./dir/dir1/emptyFile", Size: 0},
				{LocalPath: "./dir/dir1/file1", Size: 4},
				{LocalPath: "./dir/dir1/link", Size: -1, SymPath: "file1"},
			},
		}
		err = sd.TouchAll()
		So(err, ShouldBeNil)
		fd, err := os.OpenFile("./dir/dir1/file1", os.O_WRONLY, 0644)
		So(err, ShouldBeNil)
		n, err := fd.Write([]byte("1234"))
		So(err, ShouldBeNil)
		So(n, ShouldEqual, 4)
		err = fd.Close()
		So(err, ShouldBeNil)

		sd2, err := MakeSeed("./dir", 10)
		So(err, ShouldBeNil)
		for i := range sd2.Files {
			So(sd2.Files[i].Path, ShouldResemble, path.Clean(sd.Files[i].LocalPath))
			So(sd2.Files[i].Size, ShouldResemble, sd.Files[i].Size)
		}
	})
}
