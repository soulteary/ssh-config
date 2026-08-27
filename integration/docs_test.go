package integration_test

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

var (
	inlineMarkdownLink = regexp.MustCompile(`!?\[[^\]]*\]\(([^)]+)\)`)
	referenceLink      = regexp.MustCompile(`(?m)^\s*\[[^\]]+\]:\s*(\S+)`)
)

func TestDocumentationLinks(t *testing.T) {
	t.Parallel()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate documentation test")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(filename), ".."))
	paths := []string{
		filepath.Join(root, "README.md"),
		filepath.Join(root, "README_CN.md"),
		filepath.Join(root, "SECURITY.md"),
		filepath.Join(root, "CONTRIBUTING.md"),
	}
	if err := filepath.WalkDir(filepath.Join(root, "docs"), func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.EqualFold(filepath.Ext(path), ".md") {
			paths = append(paths, path)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("read %s: %v", filepath.ToSlash(path), err)
			continue
		}
		matches := inlineMarkdownLink.FindAllSubmatch(content, -1)
		matches = append(matches, referenceLink.FindAllSubmatch(content, -1)...)
		for _, match := range matches {
			checkDocumentationLink(t, root, path, string(match[1]))
		}
	}
}

func checkDocumentationLink(t *testing.T, root, source, destination string) {
	t.Helper()
	destination = strings.TrimSpace(destination)
	if strings.HasPrefix(destination, "<") {
		if end := strings.IndexByte(destination, '>'); end >= 0 {
			destination = destination[1:end]
		}
	} else if separator := strings.IndexAny(destination, " \t"); separator >= 0 {
		destination = destination[:separator]
	}
	if anchor := strings.IndexByte(destination, '#'); anchor >= 0 {
		destination = destination[:anchor]
	}
	lower := strings.ToLower(destination)
	if destination == "" || strings.HasPrefix(lower, "http://") ||
		strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "mailto:") {
		return
	}

	var target string
	if strings.HasPrefix(destination, "/") {
		target = filepath.Join(root, filepath.FromSlash(strings.TrimPrefix(destination, "/")))
	} else {
		target = filepath.Join(filepath.Dir(source), filepath.FromSlash(destination))
	}
	if _, err := os.Stat(filepath.Clean(target)); err != nil {
		relativeSource, _ := filepath.Rel(root, source)
		t.Errorf("%s links to missing local path %q", filepath.ToSlash(relativeSource), destination)
	}
}
