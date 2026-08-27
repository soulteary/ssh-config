//go:build !unix

package fn

import "os"

func openConfigFile(path string) (*os.File, error) {
	return os.Open(path)
}
