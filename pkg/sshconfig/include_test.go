package sshconfig

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestResolveIncludesLexicalOrderAndContext(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	includeDirectory := filepath.Join(directory, "config.d")
	if err := os.Mkdir(includeDirectory, 0700); err != nil {
		t.Fatal(err)
	}
	writeTestConfig(t, filepath.Join(directory, "config"), "Host example\n  Include config.d/*\nMatch host internal\n  Include missing/*\n")
	writeTestConfig(t, filepath.Join(includeDirectory, "20-second"), "Host second\n")
	writeTestConfig(t, filepath.Join(includeDirectory, "10-first"), "Host first\n")

	graph, err := ResolveIncludes(filepath.Join(directory, "config"), ResolveOptions{RelativeBase: directory})
	if err != nil {
		t.Fatalf("ResolveIncludes() error = %v", err)
	}
	wantOrder := []string{
		filepath.Join(directory, "config"),
		filepath.Join(includeDirectory, "10-first"),
		filepath.Join(includeDirectory, "20-second"),
	}
	if !reflect.DeepEqual(graph.Order, wantOrder) {
		t.Fatalf("order = %v, want %v", graph.Order, wantOrder)
	}
	if len(graph.Edges) != 2 || len(graph.Edges[1].Targets) != 0 {
		t.Fatalf("edges = %#v", graph.Edges)
	}
	entry := graph.Files[graph.Entry].Document
	got, _ := entry.MarshalPreserve()
	want, _ := os.ReadFile(graph.Entry)
	if !reflect.DeepEqual(got, want) {
		t.Fatal("resolving includes changed the entry document")
	}
}

func TestResolveIncludesExpansions(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	sshDirectory := filepath.Join(directory, ".ssh")
	if err := os.Mkdir(sshDirectory, 0700); err != nil {
		t.Fatal(err)
	}
	entry := filepath.Join(sshDirectory, "config")
	writeTestConfig(t, entry, "Include ~/hosts/${PROFILE}/%h.conf\n")
	targetDirectory := filepath.Join(directory, "hosts", "work")
	if err := os.MkdirAll(targetDirectory, 0700); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(targetDirectory, "server.conf")
	writeTestConfig(t, target, "Host server\n")

	graph, err := ResolveIncludes(entry, ResolveOptions{
		UserHome:    directory,
		Environment: map[string]string{"PROFILE": "work"},
		Tokens:      map[byte]string{'h': "server"},
	})
	if err != nil {
		t.Fatalf("ResolveIncludes() error = %v", err)
	}
	if len(graph.Order) != 2 || graph.Order[1] != target {
		t.Fatalf("order = %v", graph.Order)
	}
}

func TestResolveIncludesCycleAndDepth(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	first := filepath.Join(directory, "first")
	second := filepath.Join(directory, "second")
	writeTestConfig(t, first, "Include second\n")
	writeTestConfig(t, second, "Include first\n")
	_, err := ResolveIncludes(first, ResolveOptions{RelativeBase: directory})
	if err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("cycle error = %v", err)
	}

	writeTestConfig(t, second, "Host second\n")
	_, err = ResolveIncludes(first, ResolveOptions{RelativeBase: directory, MaxDepth: 0})
	if err != nil {
		t.Fatalf("default depth error = %v", err)
	}
	_, err = ResolveIncludes(first, ResolveOptions{RelativeBase: directory, MaxDepth: -1})
	if err != nil {
		t.Fatalf("negative default depth error = %v", err)
	}

	third := filepath.Join(directory, "third")
	writeTestConfig(t, second, "Include third\n")
	writeTestConfig(t, third, "Host third\n")
	_, err = ResolveIncludes(first, ResolveOptions{RelativeBase: directory, MaxDepth: 1})
	if err == nil || !strings.Contains(err.Error(), "depth") {
		t.Fatalf("depth error = %v", err)
	}
}

func TestResolveIncludesPermissionCheck(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits are not portable to Windows")
	}
	t.Parallel()
	directory := t.TempDir()
	entry := filepath.Join(directory, "config")
	writeTestConfig(t, entry, "Host example\n")
	if err := os.Chmod(entry, 0666); err != nil {
		t.Fatal(err)
	}
	_, err := ResolveIncludes(entry, ResolveOptions{CheckPermissions: true})
	if err == nil || !strings.Contains(err.Error(), "bad permissions") {
		t.Fatalf("permission error = %v", err)
	}
}

func TestResolveIncludesPermissionCheckReadsTheOpenedFile(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	entry := filepath.Join(directory, "config")
	writeTestConfig(t, entry, "Host example\n")

	file, err := os.Open(entry)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := checkIncludePermissions(file, entry); err != nil {
		t.Fatalf("checkIncludePermissions() error = %v", err)
	}
}

func TestResolveIncludesExpansionErrors(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	entry := filepath.Join(directory, "config")
	writeTestConfig(t, entry, "Include ${MISSING}/%x\n")
	_, err := ResolveIncludes(entry, ResolveOptions{Environment: map[string]string{}, Tokens: map[byte]string{}})
	if err == nil || !strings.Contains(err.Error(), "MISSING") {
		t.Fatalf("environment error = %v", err)
	}
	writeTestConfig(t, entry, "Include %x\n")
	_, err = ResolveIncludes(entry, ResolveOptions{Tokens: map[byte]string{}})
	if err == nil || !strings.Contains(err.Error(), "%x") {
		t.Fatalf("token error = %v", err)
	}
}

func TestResolveIncludesResourceLimits(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	entry := filepath.Join(directory, "config")
	target := filepath.Join(directory, "included")
	entryContent := "Include included\n"
	targetContent := "Host included\n"
	writeTestConfig(t, entry, entryContent)
	writeTestConfig(t, target, targetContent)

	tests := []struct {
		name    string
		options ResolveOptions
		wantErr string
	}{
		{name: "file count", options: ResolveOptions{RelativeBase: directory, MaxFiles: 1}, wantErr: "file count exceeds"},
		{name: "single file bytes", options: ResolveOptions{RelativeBase: directory, MaxFileBytes: int64(len(entryContent) - 1)}, wantErr: "exceeds maximum size"},
		{name: "total bytes", options: ResolveOptions{RelativeBase: directory, MaxTotalBytes: int64(len(entryContent) + len(targetContent) - 1)}, wantErr: "total limit"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ResolveIncludes(entry, test.options)
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("ResolveIncludes() error = %v, want %q", err, test.wantErr)
			}
		})
	}

	graph, err := ResolveIncludes(entry, ResolveOptions{
		RelativeBase:  directory,
		MaxFiles:      2,
		MaxFileBytes:  int64(len(entryContent)),
		MaxTotalBytes: int64(len(entryContent) + len(targetContent)),
	})
	if err != nil {
		t.Fatalf("exact resource limits: %v", err)
	}
	if len(graph.Order) != 2 {
		t.Fatalf("order = %v", graph.Order)
	}
}

func writeTestConfig(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}
