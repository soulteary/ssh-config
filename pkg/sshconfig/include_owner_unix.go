//go:build unix

package sshconfig

import (
	"fmt"
	"io/fs"
	"os"
	"syscall"
)

func openIncludeFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
}

func checkIncludePlatformPermissions(info fs.FileInfo, path string) error {
	if info.Mode().Perm()&0022 != 0 {
		return fmt.Errorf("sshconfig: bad permissions on %s: mode %o is writable by group or others", path, info.Mode().Perm())
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("sshconfig: inspect owner of included file %s", path)
	}
	uid := uint32(os.Getuid())
	if stat.Uid != 0 && stat.Uid != uid {
		return fmt.Errorf("sshconfig: bad owner on %s: uid %d is neither root nor current user %d", path, stat.Uid, uid)
	}
	return nil
}
