//go:build !unix

package sshconfig

import (
	"io/fs"
	"os"
)

func openIncludeFile(path string) (*os.File, error) {
	return os.Open(path)
}

// Non-Unix platforms do not expose OpenSSH's uid ownership and mode model.
func checkIncludePlatformPermissions(fs.FileInfo, string) error {
	return nil
}
