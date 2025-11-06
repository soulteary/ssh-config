package fn

import "os"

func isFileReadable(info os.FileInfo) bool {
	perm := info.Mode().Perm()
	return perm&0400 != 0 || perm&0040 != 0 || perm&0004 != 0
}

func isDirReadable(info os.FileInfo) bool {
	perm := info.Mode().Perm()
	user := perm&0500 == 0500
	group := perm&0050 == 0050
	other := perm&0005 == 0005
	return user || group || other
}

func isDirWritable(info os.FileInfo) bool {
	perm := info.Mode().Perm()
	user := perm&0300 == 0300
	group := perm&0030 == 0030
	other := perm&0003 == 0003
	return user || group || other
}
