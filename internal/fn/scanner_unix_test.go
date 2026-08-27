//go:build unix

package fn

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestScannerRejectsFIFOWithoutBlocking(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	fifo := filepath.Join(directory, "config.fifo")
	if err := syscall.Mkfifo(fifo, 0600); err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{}, 1)
	go func() {
		if IsConfigFile(fifo) {
			t.Error("FIFO was identified as an SSH config file")
		}
		if config := ReadSingleConfig(fifo); config != nil {
			t.Errorf("ReadSingleConfig(FIFO) = %#v, want nil", config)
		}
		if _, err := ReadSSHConfigs(fifo); err == nil {
			t.Error("ReadSSHConfigs(FIFO) succeeded")
		}
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("scanner blocked while opening a FIFO")
	}
}

func TestDirectoryScanSkipsFIFO(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	fifo := filepath.Join(directory, "blocked")
	if err := syscall.Mkfifo(fifo, 0600); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(directory, "config")
	if err := os.WriteFile(configPath, []byte("Host example\n"), 0600); err != nil {
		t.Fatal(err)
	}

	done := make(chan *SSHConfig, 1)
	errs := make(chan error, 1)
	go func() {
		config, err := ReadSSHConfigs(directory)
		if err != nil {
			errs <- err
			return
		}
		done <- config
	}()

	select {
	case err := <-errs:
		t.Fatal(err)
	case config := <-done:
		if _, ok := config.Configs[configPath]; !ok {
			t.Fatalf("regular config was not scanned: %#v", config.Configs)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("directory scan blocked on a FIFO")
	}
}
