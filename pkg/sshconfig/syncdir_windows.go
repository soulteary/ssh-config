//go:build windows

package sshconfig

// Windows does not support syncing a directory handle opened through os.Open.
// The temporary file itself is synced before the atomic rename.
func syncDirectory(string) error {
	return nil
}
