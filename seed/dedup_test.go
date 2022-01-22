package seed

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSortDeDup(t *testing.T) {
	Convey("sort and deduplicate", t, func() {
		l := []Host{
			{[4]byte{10, 0, 0, 1}, 10},
		}
		sl := SortDeDup(l)
		So(sl, ShouldResemble, []Host{
			{[4]byte{10, 0, 0, 1}, 10},
		})
	})
	Convey("sort and deduplicate", t, func() {
		l := []Host{
			{[4]byte{10, 0, 0, 1}, 10},
			{[4]byte{10, 0, 0, 1}, 10},
			{[4]byte{10, 0, 0, 2}, 10},
			{[4]byte{10, 0, 0, 1}, 8},
		}
		sl := SortDeDup(l)
		So(sl, ShouldResemble, []Host{
			{[4]byte{10, 0, 0, 1}, 8},
			{[4]byte{10, 0, 0, 1}, 10},
			{[4]byte{10, 0, 0, 2}, 10},
		})
	})
}
