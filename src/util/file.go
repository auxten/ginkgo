package util

import "os"

func IsFile(m os.FileMode) bool{
	return m & os.ModeType == 0
}

func IsDir(m os.FileMode) bool{
	return m & os.ModeDir == 1
}
