package sshconfig

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// MarshalPreserve serializes the document while copying unchanged nodes
// directly from the original source.
func (d *Document) MarshalPreserve() ([]byte, error) {
	if d == nil {
		return nil, fmt.Errorf("sshconfig: nil document")
	}
	var out bytes.Buffer
	for _, node := range d.nodes {
		if d.removed[node.ID] {
			continue
		}
		if replacement, ok := d.replacement[node.ID]; ok {
			out.Write(replacement)
			continue
		}
		if node.Span.Start < 0 || node.Span.End < node.Span.Start || node.Span.End > len(d.source) {
			return nil, fmt.Errorf("sshconfig: invalid span for node %d", node.ID)
		}
		out.Write(d.source[node.Span.Start:node.Span.End])
	}
	return out.Bytes(), nil
}

// SaveOptions controls atomic file replacement.
type SaveOptions struct {
	// Mode is used for a new file. Zero defaults to 0600.
	Mode fs.FileMode
	// PreserveMode retains an existing regular file's permission bits.
	PreserveMode bool
}

// Save atomically writes a document to path. Symbolic-link destinations are
// rejected so a caller cannot unintentionally replace a link target.
func (d *Document) Save(path string, options SaveOptions) error {
	data, err := d.MarshalPreserve()
	if err != nil {
		return err
	}
	return SaveAtomic(path, data, options)
}

// SaveAtomic writes data to a temporary file in the destination directory,
// syncs it, and atomically renames it over path.
func SaveAtomic(path string, data []byte, options SaveOptions) error {
	if path == "" {
		return fmt.Errorf("sshconfig: destination path is empty")
	}
	mode := options.Mode.Perm()
	if mode == 0 {
		mode = 0600
	}
	if info, err := os.Lstat(path); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("sshconfig: refusing symbolic-link destination %s", path)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("sshconfig: destination %s is not a regular file", path)
		}
		if options.PreserveMode {
			mode = info.Mode().Perm()
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("sshconfig: inspect destination %s: %w", path, err)
	}

	directory := filepath.Dir(path)
	temp, err := os.CreateTemp(directory, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("sshconfig: create temporary file: %w", err)
	}
	tempPath := temp.Name()
	closed := false
	defer func() {
		if !closed {
			_ = temp.Close()
		}
		_ = os.Remove(tempPath)
	}()

	if err := temp.Chmod(mode); err != nil {
		return fmt.Errorf("sshconfig: set temporary file mode: %w", err)
	}
	if _, err := temp.Write(data); err != nil {
		return fmt.Errorf("sshconfig: write temporary file: %w", err)
	}
	if err := temp.Sync(); err != nil {
		return fmt.Errorf("sshconfig: sync temporary file: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("sshconfig: close temporary file: %w", err)
	}
	closed = true
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("sshconfig: replace destination %s: %w", path, err)
	}
	return nil
}
