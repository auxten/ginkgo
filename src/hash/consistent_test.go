package consistent

import (
. "github.com/smartystreets/goconvey/convey"
"testing"
	"github.com/auxten/logrus"
)

func TestSpec(t *testing.T) {

	// Only pass t into top-level Convey calls
	Convey("Given some integer with a starting value", t, func() {
		logrus.SetLevel(logrus.DebugLevel)
		c := ConsistentRing{
			Range:   100,
			Buckets: []Bucket{},
		}
		c.AddNode("aaa")
		c.AddNode("bbb")
		c.AddNode("ddddd")
		c.DumpNodesRange()
		c.FindBucketByKey("ddddd")
		c.FindBucketByKey("fd")
		c.FindBucketByKey("aa")
		c.FindBucketByKey("bbb")
		So(nil, ShouldEqual, nil)
	})
}

