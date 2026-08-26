//go:build !windows

package sshconfig

import "os"

// syncDirectory makes the rename durable across a crash on filesystems that
// require the containing directory entry to be flushed separately.
func syncDirectory(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}
