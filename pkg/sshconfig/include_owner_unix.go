//go:build unix

package sshconfig

import (
	"fmt"
	"io/fs"
	"os"
	"syscall"
)

func checkIncludeOwner(info fs.FileInfo, path string) error {
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
