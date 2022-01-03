package seed

import (
	"encoding/json"
	"fmt"
	"testing"

	log "github.com/auxten/logrus"
)

func TestMakeSeed(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	var testData = []struct {
		blockSize int64
		js        string
	}{
		{-1, `{"fileCount":10,"files":[{"mode":2147484141,"size":-1,"symPath":"","path":"../test","checkSum":null},{"mode":2147484141,"size":-1,"symPath":"","path":"../test/dir1","checkSum":null},{"mode":420,"size":0,"symPath":"","path":"../test/dir1/emptyFile","checkSum":null},{"mode":420,"size":9,"symPath":"","path":"../test/dir1/file11","checkSum":null},{"mode":420,"size":21,"symPath":"","path":"../test/dir1/file12","checkSum":null},{"mode":420,"size":13,"symPath":"","path":"../test/dir1/file13","checkSum":null},{"mode":2147484141,"size":-1,"symPath":"","path":"../test/dir2","checkSum":null},{"mode":2147484141,"size":-1,"symPath":"","path":"../test/dir2/dir21","checkSum":null},{"mode":134218221,"size":-2,"symPath":"dir1/file11","path":"../test/dir2/dir21/ln211","checkSum":null},{"mode":2147484141,"size":-1,"symPath":"","path":"../test/emptyDir","checkSum":null}],"blocks":[{"size":43,"startFile":3,"startOffset":0,"checkSum":null}],"blockSize":4194304,"totalSize":43}`},
		{10, `{"fileCount":10,"files":[{"mode":2147484141,"size":-1,"symPath":"","path":"../test","checkSum":null},{"mode":2147484141,"size":-1,"symPath":"","path":"../test/dir1","checkSum":null},{"mode":420,"size":0,"symPath":"","path":"../test/dir1/emptyFile","checkSum":null},{"mode":420,"size":9,"symPath":"","path":"../test/dir1/file11","checkSum":null},{"mode":420,"size":21,"symPath":"","path":"../test/dir1/file12","checkSum":null},{"mode":420,"size":13,"symPath":"","path":"../test/dir1/file13","checkSum":null},{"mode":2147484141,"size":-1,"symPath":"","path":"../test/dir2","checkSum":null},{"mode":2147484141,"size":-1,"symPath":"","path":"../test/dir2/dir21","checkSum":null},{"mode":134218221,"size":-2,"symPath":"dir1/file11","path":"../test/dir2/dir21/ln211","checkSum":null},{"mode":2147484141,"size":-1,"symPath":"","path":"../test/emptyDir","checkSum":null}],"blocks":[{"size":10,"startFile":3,"startOffset":0,"checkSum":null},{"size":10,"startFile":4,"startOffset":1,"checkSum":null},{"size":10,"startFile":4,"startOffset":11,"checkSum":null},{"size":10,"startFile":5,"startOffset":0,"checkSum":null},{"size":3,"startFile":5,"startOffset":10,"checkSum":null}],"blockSize":10,"totalSize":43}`},
	}
	for i, d := range testData {
		t.Run(fmt.Sprintf("make seed %d", i), func(t *testing.T) {

			seed, err := MakeSeed("../test", d.blockSize)
			if err != nil {
				t.Error(err)
			}
			jb, err := json.Marshal(seed)
			if err != nil {
				t.Error(err)
			}
			if string(jb) != d.js {
				t.Errorf("Expected %s, got %s", d.js, string(jb))
			}
		})
	}

}
