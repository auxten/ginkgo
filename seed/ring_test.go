package seed

import (
	"sync"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLocator(t *testing.T) {
	Convey("consistent hash ring", t, func() {
		ring := &Seed{
			RWMutex:    sync.RWMutex{},
			Blocks:     []*Block{{}, {}, {}, {}, {}, {}, {}, {}},
			VNodeCount: 2,
		}
		ring.Remove(Host{
			IP:   [4]byte{},
			Port: 0,
		})
		hosts := ring.LocateBlock(1, 1)
		So(hosts, ShouldHaveLength, 0)
		hosts = ring.LocateBlock(1000, 1)
		So(hosts, ShouldHaveLength, 0)
		ring.Add(Host{
			IP:   [4]byte{},
			Port: 0,
		})
		hosts = ring.LocateBlock(1, 1)
		So(hosts, ShouldHaveLength, 1)
		hosts = ring.LocateBlock(1000, 1)
		So(hosts, ShouldHaveLength, 1)
		So(hosts[0].String(), ShouldResemble, "0.0.0.0:0")
		So(ring.GetAllHosts()[0].String(), ShouldResemble, "0.0.0.0:0")

		hosts = ring.LocateBlock(1, 2)
		So(hosts, ShouldHaveLength, 1)
		hosts = ring.LocateBlock(1000, 2)
		So(hosts, ShouldHaveLength, 1)
		So(hosts[0].String(), ShouldResemble, "0.0.0.0:0")

		ring.Remove(Host{
			IP:   [4]byte{},
			Port: 0,
		})
		hosts = ring.LocateBlock(1, 1)
		So(hosts, ShouldHaveLength, 0)
		hosts = ring.LocateBlock(1000, 1)
		So(hosts, ShouldHaveLength, 0)
		So(ring.GetAllHosts(), ShouldHaveLength, 0)

		for i := 0; i < 1000; i++ {
			ring.Add(Host{
				IP:   [4]byte{10, 0, byte(i / 256), byte(i % 256)},
				Port: uint16(i),
			})
		}

		hosts = ring.LocateBlock(1, 2)
		So(hosts, ShouldHaveLength, 2)
		hosts = ring.LocateBlock(1000, 2)
		So(hosts, ShouldHaveLength, 2)
		So(hosts[0].String(), ShouldNotResemble, hosts[1].String())
		hosts2 := ring.LocateBlock(1, 2)
		// should not deterministic
		So(hosts[0].String(), ShouldNotResemble, hosts2[0].String())
		So(ring.GetAllHosts(), ShouldHaveLength, 1000)

		for i := 0; i < 1000; i++ {
			ring.Remove(Host{
				IP:   [4]byte{10, 0, byte(i / 256), byte(i % 256)},
				Port: uint16(i),
			})
		}
		hosts = ring.LocateBlock(1, 1)
		So(hosts, ShouldHaveLength, 0)
		hosts = ring.LocateBlock(1000, 1)
		So(hosts, ShouldHaveLength, 0)
		So(ring.GetAllHosts(), ShouldHaveLength, 0)
	})
}

func TestParseHost(t *testing.T) {
	Convey("parse IPv4 host", t, func() {
		h, err := ParseHost("127.0.0.1:10001")
		So(err, ShouldBeNil)
		So(h, ShouldResemble, Host{
			IP:   [4]byte{127, 0, 0, 1},
			Port: 10001,
		})
	})
	Convey("parse IPv6 host", t, func() {
		h, err := ParseHost("[::1]:10001")
		So(err, ShouldNotBeNil)
		So(h, ShouldResemble, Host{
			IP:   [4]byte{0, 0, 0, 0},
			Port: 0,
		})
	})
	Convey("parse bad host", t, func() {
		h, err := ParseHost("127.0.0.1:1000001")
		So(err, ShouldNotBeNil)
		So(h, ShouldResemble, Host{
			IP:   [4]byte{0, 0, 0, 0},
			Port: 0,
		})
	})
	Convey("parse bad host", t, func() {
		h, err := ParseHost("1270.0.1:1000001")
		So(err, ShouldNotBeNil)
		So(h, ShouldResemble, Host{
			IP:   [4]byte{0, 0, 0, 0},
			Port: 0,
		})
	})
}
