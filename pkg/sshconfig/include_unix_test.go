//go:build unix

package sshconfig

import (
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestResolveIncludesRejectsFIFOWithoutBlocking(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "config.fifo")
	if err := syscall.Mkfifo(path, 0600); err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() {
		_, err := ResolveIncludes(path, ResolveOptions{CheckPermissions: true})
		done <- err
	}()

	select {
	case err := <-done:
		if err == nil || !strings.Contains(err.Error(), "not a regular file") {
			t.Fatalf("FIFO error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("ResolveIncludes blocked while opening a FIFO")
	}
}
