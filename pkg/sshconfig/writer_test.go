package sshconfig

import (
	"bytes"
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
