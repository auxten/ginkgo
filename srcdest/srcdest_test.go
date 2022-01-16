package srcdest

import (
	"fmt"
	"testing"
)

func TestName(t *testing.T) {
	var testData = []struct {
		cSrc, cDest         string
		cSrcType, cDestType PathType
		src, dest           string
	}{
		{"./src", "./dest", FileType, NotExist, "./src", "dest"},
		{"./src", "./dest", FileType, FileType, "./src", "dest"},
		{"./src", "./dest", FileType, DirType, "./src", "dest/src"},
		{"./src", "./dest/", FileType, NotExist, "./src", ""}, //error
		{"./src", "./dest/", FileType, FileType, "./src", ""}, //error
		{"./src", "./dest/", FileType, DirType, "./src", "dest/src"},
		{"./src", "./dest", DirType, NotExist, "./src/file", "dest/file"},
		{"./src", "./dest", DirType, FileType, "./src/file", ""}, //error
		{"./src", "./dest", DirType, DirType, "./src/file", "dest/src/file"},
		{"./src", "./dest/", DirType, NotExist, "./src/file", "dest/file"},
		{"./src", "./dest/", DirType, FileType, "./src/file", ""}, //error
		{"./src", "./dest/", DirType, DirType, "./src/file", "dest/src/file"},
		// Long path cases
		{"./srcDir/src", "./destDir/dest", FileType, NotExist, "./srcDir/src", "destDir/dest"},
		{"./src", "./dest", DirType, NotExist, "./src/srcDir/file", "dest/srcDir/file"},
		{"./src", "./xxx/../dest", DirType, NotExist, "./src/srcDir/file", "dest/srcDir/file"},
		{"./src", "./destDir/dest", DirType, NotExist, "./src/srcDir/file", "destDir/dest/srcDir/file"},
		// Other error cases
		{"./src", "", FileType, DirType, "./notSrc", ""},
	}
	for i, d := range testData {
		t.Run(fmt.Sprintf("test path normalization %d", i), func(t *testing.T) {
			path, _ := NormalizeDestPath(d.cSrc, d.cDest, d.cSrcType, d.cDestType, d.src)
			if path != d.dest {
				t.Errorf("expect: %s, got: %s", d.dest, path)
			}
		})
	}

}
