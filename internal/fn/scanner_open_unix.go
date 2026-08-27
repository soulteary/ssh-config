//go:build unix

package fn

import (
	"os"
	"syscall"
)

func openConfigFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
}
