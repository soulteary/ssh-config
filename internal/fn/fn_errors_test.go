package fn

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetPathContentReadFileError(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config")
	if err := os.WriteFile(configPath, []byte("Host example\n  HostName example.com\n"), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	originalReadFile := readFile
	readFile = func(path string) ([]byte, error) {
		if path == configPath {
			return nil, errors.New("read failure")
		}
		return originalReadFile(path)
	}
	t.Cleanup(func() { readFile = originalReadFile })

	_, err := GetPathContent(tmpDir)
	if err == nil {
		t.Fatalf("expected error but got nil")
	}
	if !strings.Contains(err.Error(), "no valid SSH config found") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestSaveStatError(t *testing.T) {
	tmpDir := t.TempDir()
	dest := filepath.Join(tmpDir, "sub", "file.txt")
	destDir := filepath.Dir(dest)

	callCount := 0
	originalStat := stat
	stat = func(path string) (os.FileInfo, error) {
		if path == destDir {
			callCount++
			if callCount == 2 {
				return nil, errors.New("stat failure")
			}
		}
		return originalStat(path)
	}
	t.Cleanup(func() { stat = originalStat })

	err := Save(dest, []byte("content"))
	if err == nil {
		t.Fatalf("expected error but got nil")
	}
	if !strings.Contains(err.Error(), "can not write to destination file") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestSaveWriteFileError(t *testing.T) {
	tmpDir := t.TempDir()
	dest := filepath.Join(tmpDir, "file.txt")

	originalWriteFile := writeFile
	writeFile = func(path string, data []byte, perm os.FileMode) error {
		if path == dest {
			return errors.New("write failure")
		}
		return originalWriteFile(path, data, perm)
	}
	t.Cleanup(func() { writeFile = originalWriteFile })

	err := Save(dest, []byte("content"))
	if err == nil {
		t.Fatalf("expected error but got nil")
	}
	if !strings.Contains(err.Error(), "can not write to destination file") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestEnsureDirectoryExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	destDir := filepath.Join(tmpDir, "existing")

	if err := os.WriteFile(destDir, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	err := ensureDirectory(destDir)
	if err == nil {
		t.Fatalf("expected error but got nil")
	}
	if !strings.Contains(err.Error(), "is not a directory") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestEnsureDirectoryParentNotDir(t *testing.T) {
	tmpDir := t.TempDir()
	parent := filepath.Join(tmpDir, "parent")
	if err := os.WriteFile(parent, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create parent file: %v", err)
	}
	destDir := filepath.Join(parent, "child")

	originalStat := stat
	stat = func(path string) (os.FileInfo, error) {
		if path == destDir {
			return nil, os.ErrNotExist
		}
		return originalStat(path)
	}
	t.Cleanup(func() { stat = originalStat })

	err := ensureDirectory(destDir)
	if err == nil {
		t.Fatalf("expected error but got nil")
	}
	if !strings.Contains(err.Error(), "parent") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestEnsureDirectoryMkdirError(t *testing.T) {
	tmpDir := t.TempDir()
	destDir := filepath.Join(tmpDir, "newdir")

	originalMkdirAll := mkdirAll
	mkdirAll = func(path string, perm os.FileMode) error {
		if path == destDir {
			return errors.New("mkdir failure")
		}
		return originalMkdirAll(path, perm)
	}
	t.Cleanup(func() { mkdirAll = originalMkdirAll })

	err := ensureDirectory(destDir)
	if err == nil {
		t.Fatalf("expected error but got nil")
	}
	if !strings.Contains(err.Error(), "can not create destination directory") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestEnsureDirectoryStatUnexpectedError(t *testing.T) {
	tmpDir := t.TempDir()
	destDir := filepath.Join(tmpDir, "unexpected")

	originalStat := stat
	stat = func(path string) (os.FileInfo, error) {
		if path == destDir {
			return nil, os.ErrPermission
		}
		return originalStat(path)
	}
	t.Cleanup(func() { stat = originalStat })

	err := ensureDirectory(destDir)
	if err == nil {
		t.Fatalf("expected error but got nil")
	}
	if !strings.Contains(err.Error(), "can not create destination directory") {
		t.Fatalf("unexpected error message: %v", err)
	}
}
