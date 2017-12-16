package seed

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	log "github.com/auxten/logrus"
)

func TestSpec(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	Convey("Seed Scan", t, func() {
		const root_path = "../"
		seed := Seed{}
		err := seed.MakeSeed(root_path)
		So(err, ShouldEqual, nil)
	})
}
