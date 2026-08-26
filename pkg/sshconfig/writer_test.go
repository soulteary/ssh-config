package sshconfig

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestMarshalPreserveWithoutEdits(t *testing.T) {
	t.Parallel()
	input := []byte("# comment\r\nHost=example\r\n\tUser root\r\n")
	doc, _ := Parse(input)
	got, err := doc.MarshalPreserve()
	if err != nil {
		t.Fatalf("MarshalPreserve() error = %v", err)
	}
	if !bytes.Equal(got, input) {
		t.Fatalf("output = %q, want %q", got, input)
	}
}

func TestReplaceDirectivePreservesSurroundings(t *testing.T) {
	t.Parallel()
	input := []byte("Host example\r\n\tUser = old # account\r\nIdentityFile first\r\nIdentityFile second\r\n")
	doc, _ := Parse(input)
	if err := doc.ReplaceDirective(1, "User", "new user"); err != nil {
		t.Fatalf("ReplaceDirective() error = %v", err)
	}
	got, err := doc.MarshalPreserve()
	if err != nil {
		t.Fatalf("MarshalPreserve() error = %v", err)
	}
	want := []byte("Host example\r\n\tUser = \"new user\" # account\r\nIdentityFile first\r\nIdentityFile second\r\n")
	if !bytes.Equal(got, want) {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestRemoveAndInsertDirective(t *testing.T) {
	t.Parallel()
	doc, _ := Parse([]byte("Host example\nIdentityFile first"))
	if err := doc.RemoveNode(1); err != nil {
		t.Fatalf("RemoveNode() error = %v", err)
	}
	id, err := doc.InsertDirectiveAfter(0, "IdentityFile", "second")
	if err != nil {
		t.Fatalf("InsertDirectiveAfter() error = %v", err)
	}
	if id < 2 {
		t.Fatalf("inserted id = %d", id)
	}
	if _, err := doc.AppendDirective("SetEnv", "FOO=bar", "HASH=#value"); err != nil {
		t.Fatalf("AppendDirective() error = %v", err)
	}
	got, _ := doc.MarshalPreserve()
	want := "Host example\nIdentityFile second\nSetEnv FOO=bar \"HASH=#value\"\n"
	if string(got) != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestEditErrors(t *testing.T) {
	t.Parallel()
	doc, _ := Parse([]byte("# comment\n"))
	if err := doc.ReplaceDirective(0, "User", "root"); err == nil {
		t.Fatal("replaced a comment node")
	}
	if err := doc.RemoveNode(99); err == nil {
		t.Fatal("removed an unknown node")
	}
	if _, err := doc.InsertDirectiveAfter(99, "User", "root"); err == nil {
		t.Fatal("inserted after an unknown node")
	}
	if _, err := doc.AppendDirective(""); err == nil {
		t.Fatal("appended an empty keyword")
	}
}

func TestDocumentEditsRejectCrossLineInput(t *testing.T) {
	doc, _ := Parse([]byte("Host example\n"))
	tests := []struct {
		name string
		edit func() error
	}{
		{name: "keyword newline", edit: func() error { return doc.ReplaceDirective(0, "Host\nProxyCommand", "example") }},
		{name: "keyword comment", edit: func() error { return doc.ReplaceDirective(0, "#Host", "example") }},
		{name: "argument newline", edit: func() error { return doc.ReplaceDirective(0, "Host", "safe\nProxyCommand command") }},
		{name: "argument NUL", edit: func() error { _, err := doc.AppendDirective("Host", "safe\x00hidden"); return err }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.edit(); err == nil {
				t.Fatal("edit accepted input that cannot fit in one physical directive")
			}
		})
	}
}

func TestSaveAtomic(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	path := filepath.Join(directory, "config")
	if err := os.WriteFile(path, []byte("old"), 0640); err != nil {
		t.Fatal(err)
	}
	doc, _ := Parse([]byte("Host example\n"))
	if err := doc.Save(path, SaveOptions{PreserveMode: true}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "Host example\n" {
		t.Fatalf("saved content = %q", got)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if gotMode := info.Mode().Perm(); gotMode != 0640 {
		t.Fatalf("saved mode = %o, want 640", gotMode)
	}
	matches, err := filepath.Glob(filepath.Join(directory, ".config.tmp-*"))
	if err != nil || len(matches) != 0 {
		t.Fatalf("temporary files remain: %v, %v", matches, err)
	}
}

func TestSaveAtomicRejectsSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symbolic links require additional privileges on Windows")
	}
	t.Parallel()
	directory := t.TempDir()
	target := filepath.Join(directory, "target")
	link := filepath.Join(directory, "config")
	if err := os.WriteFile(target, []byte("unchanged"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if err := SaveAtomic(link, []byte("changed"), SaveOptions{}); err == nil {
		t.Fatal("SaveAtomic() accepted a symbolic link")
	}
	got, err := os.ReadFile(target)
	if err != nil || string(got) != "unchanged" {
		t.Fatalf("target changed: %q, %v", got, err)
	}
}

type failingAtomicFile struct {
	*os.File
	failure string
}

func (f *failingAtomicFile) Chmod(mode fs.FileMode) error {
	if f.failure == "chmod" {
		return errors.New("injected chmod failure")
	}
	return f.File.Chmod(mode)
}

func (f *failingAtomicFile) Write(data []byte) (int, error) {
	if f.failure == "write" {
		return 0, errors.New("injected write failure")
	}
	return f.File.Write(data)
}

func (f *failingAtomicFile) Sync() error {
	if f.failure == "sync" {
		return errors.New("injected sync failure")
	}
	return f.File.Sync()
}

func (f *failingAtomicFile) Close() error {
	err := f.File.Close()
	if f.failure == "close" {
		return errors.New("injected close failure")
	}
	return err
}

func TestSaveAtomicFailureBeforeRenameKeepsDestination(t *testing.T) {
	t.Parallel()
	for _, failure := range []string{"chmod", "write", "sync", "close", "rename"} {
		failure := failure
		t.Run(failure, func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			path := filepath.Join(directory, "config")
			if err := os.WriteFile(path, []byte("old"), 0600); err != nil {
				t.Fatal(err)
			}

			operations := defaultAtomicWriteOperations
			operations.createTemp = func(directory, pattern string) (atomicFile, error) {
				file, err := os.CreateTemp(directory, pattern)
				if err != nil {
					return nil, err
				}
				return &failingAtomicFile{File: file, failure: failure}, nil
			}
			if failure == "rename" {
				operations.rename = func(string, string) error { return errors.New("injected rename failure") }
			}

			if err := saveAtomic(path, []byte("new"), SaveOptions{}, operations); err == nil {
				t.Fatal("saveAtomic() unexpectedly succeeded")
			}
			got, err := os.ReadFile(path)
			if err != nil || string(got) != "old" {
				t.Fatalf("destination = %q, %v; want unchanged", got, err)
			}
			matches, err := filepath.Glob(filepath.Join(directory, ".config.tmp-*"))
			if err != nil || len(matches) != 0 {
				t.Fatalf("temporary files remain: %v, %v", matches, err)
			}
		})
	}
}

func TestSaveAtomicReportsDirectorySyncFailure(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	path := filepath.Join(directory, "config")
	operations := defaultAtomicWriteOperations
	syncErr := errors.New("injected directory sync failure")
	operations.syncDirectory = func(string) error { return syncErr }

	err := saveAtomic(path, []byte("new"), SaveOptions{}, operations)
	if err == nil || !errors.Is(err, syncErr) {
		t.Fatalf("saveAtomic() error = %v", err)
	}
	got, readErr := os.ReadFile(path)
	if readErr != nil || string(got) != "new" {
		t.Fatalf("renamed destination = %q, %v", got, readErr)
	}
}
