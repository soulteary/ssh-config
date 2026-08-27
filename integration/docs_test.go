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
	referenceLink = regexp.MustCompile(`(?m)^\s*\[[^\]]+\]:\s*(\S+)`)
)

func TestDocumentationLinks(t *testing.T) {
	t.Parallel()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate documentation test")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(filename), ".."))
	var paths []string
	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && entry.Name() == ".git" {
			return filepath.SkipDir
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
		for _, destination := range markdownDestinations(content) {
			checkDocumentationLink(t, root, path, destination)
		}
	}
}

func markdownDestinations(content []byte) []string {
	var destinations []string
	for index := 0; index < len(content); index++ {
		if content[index] != '[' || escapedMarkdownByte(content, index) {
			continue
		}
		labelEnd, ok := matchingMarkdownDelimiter(content, index, '[', ']')
		if !ok {
			continue
		}
		destinationStart := labelEnd + 1
		for destinationStart < len(content) && (content[destinationStart] == ' ' || content[destinationStart] == '\t') {
			destinationStart++
		}
		if destinationStart >= len(content) || content[destinationStart] != '(' {
			continue
		}
		destinationEnd, ok := matchingMarkdownDelimiter(content, destinationStart, '(', ')')
		if ok {
			if destination := markdownDestination(string(content[destinationStart+1 : destinationEnd])); destination != "" {
				destinations = append(destinations, destination)
			}
		}
	}
	for _, match := range referenceLink.FindAllSubmatch(content, -1) {
		destinations = append(destinations, string(match[1]))
	}
	return destinations
}

func markdownDestination(body string) string {
	body = strings.TrimSpace(body)
	if body == "" || body[0] == '"' || body[0] == '\'' {
		return ""
	}
	if body[0] == '<' {
		if end := strings.IndexByte(body, '>'); end >= 0 {
			return body[:end+1]
		}
		return ""
	}
	if separator := strings.IndexAny(body, " \t\r\n"); separator >= 0 {
		return body[:separator]
	}
	return body
}

func matchingMarkdownDelimiter(content []byte, start int, open, close byte) (int, bool) {
	depth := 0
	quote := byte(0)
	angleDestination := false
	for index := start; index < len(content); index++ {
		if escapedMarkdownByte(content, index) {
			continue
		}
		character := content[index]
		if open == '(' {
			if quote != 0 {
				if character == quote {
					quote = 0
				}
				continue
			}
			if angleDestination {
				if character == '>' {
					angleDestination = false
				}
				continue
			}
			if depth == 1 && character == '<' {
				angleDestination = true
				continue
			}
			if depth == 1 && (character == '"' || character == '\'') {
				quote = character
				continue
			}
		}
		switch character {
		case open:
			depth++
		case close:
			depth--
			if depth == 0 {
				return index, true
			}
		}
	}
	return 0, false
}

func TestDocumentationMarkdownDestinationsIgnoreTitleAndAngleParentheses(t *testing.T) {
	t.Parallel()
	destinations := markdownDestinations([]byte(`[guide](missing.md "see (draft") [angle](<missing(two).md>)`))
	want := map[string]bool{
		`missing.md`:        true,
		`<missing(two).md>`: true,
	}
	for _, destination := range destinations {
		delete(want, destination)
	}
	if len(want) != 0 {
		t.Fatalf("markdownDestinations() missed destinations with title or angle parentheses: %v", want)
	}
}

func TestDocumentationMarkdownDestinationsSkipTitleOnlyLinks(t *testing.T) {
	t.Parallel()
	destinations := markdownDestinations([]byte(`[help]( "see (draft")`))
	if len(destinations) != 0 {
		t.Fatalf("markdownDestinations() treated a title as a destination: %v", destinations)
	}
}

func escapedMarkdownByte(content []byte, index int) bool {
	backslashes := 0
	for index--; index >= 0 && content[index] == '\\'; index-- {
		backslashes++
	}
	return backslashes%2 == 1
}

func TestDocumentationMarkdownDestinations(t *testing.T) {
	t.Parallel()
	destinations := markdownDestinations([]byte(`[![badge](./image.svg)](./report.md) [guide](./guide.md)`))
	want := map[string]bool{"./image.svg": true, "./report.md": true, "./guide.md": true}
	for _, destination := range destinations {
		delete(want, destination)
	}
	if len(want) != 0 {
		t.Fatalf("markdownDestinations() missed nested destinations: %v", want)
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
