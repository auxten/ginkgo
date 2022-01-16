package srcdest

import (
	"fmt"
	"path/filepath"
	"strings"
)

type PathType int

const (
	NotExist PathType = iota
	FileType
	DirType
)

// NormalizeDestPath get the path of destination
// all path will be cleaned, see filepath.Clean()
// for origCmdDestPath the tailing '/' will be PRESERVED
//
// for better understanding, please see the test cases
func NormalizeDestPath(cmdSrcPath, origCmdDestPath string, cmdSrcType, cmdDestType PathType, srcPath string) (destPath string, err error) {
	cmdSrcPath = filepath.Clean(cmdSrcPath)
	srcPath = filepath.Clean(srcPath)
	cmdDestPath := filepath.Clean(origCmdDestPath)
	if len(origCmdDestPath) > 1 && origCmdDestPath[len(origCmdDestPath)-1] == '/' {
		cmdDestPath = cmdDestPath + "/"
	}
	if cmdSrcPath != "." && !strings.HasPrefix(srcPath, cmdSrcPath) {
		return "", fmt.Errorf(
			"src path %s not started with cmd src path %s", srcPath, cmdSrcPath)
	}
	switch cmdSrcType {
	case FileType:
		switch cmdDestType {
		case NotExist, FileType:
			if cmdDestPath[len(cmdDestPath)-1] == '/' {
				err = fmt.Errorf("dest %s not dir", cmdDestPath)
			} else {
				destPath = filepath.Clean(cmdDestPath)
			}
			return
		case DirType:
			destPath = filepath.Join(cmdDestPath, filepath.Base(srcPath))
			return
		}
	case DirType:
		switch cmdDestType {
		case NotExist:
			var relPath string
			if relPath, err = filepath.Rel(cmdSrcPath, srcPath); err == nil {
				destPath = filepath.Join(cmdDestPath, relPath)
			}
			return
		case FileType:
			err = fmt.Errorf("copy dir %s to file %s", cmdSrcPath, cmdDestPath)
			return
		case DirType:
			destPath = filepath.Join(cmdDestPath, srcPath)
			return
		}
	}
	return
}
