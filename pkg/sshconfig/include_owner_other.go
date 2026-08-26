//go:build !unix

package sshconfig

import "io/fs"

// Non-Unix platforms do not expose OpenSSH's uid ownership model.
func checkIncludeOwner(fs.FileInfo, string) error {
	return nil
}
