package seed

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestHost_Hash(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	const (
		size   = 256
		vCount = 5
	)

	var testData = []struct {
		host Host
	}{
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
		if counter > 10 {
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
		if counter > 10 {
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
		if counter > 10 {
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
