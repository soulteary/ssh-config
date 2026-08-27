package sshconfig

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const defaultIncludeDepth = 16

// ResolveOptions controls OpenSSH Include path expansion.
type ResolveOptions struct {
	// MaxDepth defaults to OpenSSH's READCONF_MAX_DEPTH value of 16.
	MaxDepth int
	// RelativeBase is used for non-absolute Include paths. When empty,
	// UserHome/.ssh is used if UserHome is set, otherwise the entry file's
	// directory is used.
	RelativeBase string
	UserHome     string
	// Environment enables ${NAME} expansion using this explicit map. A nil
	// map disables environment expansion.
	Environment map[string]string
	// Tokens supplies OpenSSH-style percent token values, for example 'h'.
	Tokens map[byte]string
	// CheckPermissions rejects files writable by group or others.
	CheckPermissions bool
}

// ResolvedFile is one parsed file in an Include graph.
type ResolvedFile struct {
	Path     string
	Document *Document
}

// IncludeEdge records one Include directive and its lexically ordered files.
type IncludeEdge struct {
	FromPath string
	NodeID   NodeID
	Pattern  string
	Targets  []string
}

// DocumentGraph retains all source documents and Include relationships.
type DocumentGraph struct {
	Entry string
	Files map[string]*ResolvedFile
	Edges []IncludeEdge
	// Order records traversal order, including repeated non-recursive includes.
	Order []string
}

// ResolveIncludes parses entry and recursively builds its Include graph.
// Include directives are retained in their source documents.
func ResolveIncludes(entry string, options ResolveOptions) (*DocumentGraph, error) {
	if entry == "" {
		return nil, fmt.Errorf("sshconfig: include entry path is empty")
	}
	absolute, err := filepath.Abs(entry)
	if err != nil {
		return nil, fmt.Errorf("sshconfig: resolve entry path: %w", err)
	}
	absolute = filepath.Clean(absolute)
	if options.MaxDepth <= 0 {
		options.MaxDepth = defaultIncludeDepth
	}
	if options.RelativeBase == "" {
		if options.UserHome != "" {
			options.RelativeBase = filepath.Join(options.UserHome, ".ssh")
		} else {
			options.RelativeBase = filepath.Dir(absolute)
		}
	}
	graph := &DocumentGraph{
		Entry: absolute,
		Files: make(map[string]*ResolvedFile),
	}
	if err := graph.resolveFile(absolute, options, 0, make(map[string]bool)); err != nil {
		return nil, err
	}
	return graph, nil
}

func (g *DocumentGraph) resolveFile(path string, options ResolveOptions, depth int, stack map[string]bool) error {
	path = filepath.Clean(path)
	if depth > options.MaxDepth {
		return fmt.Errorf("sshconfig: include depth exceeds %d at %s", options.MaxDepth, path)
	}
	if stack[path] {
		return fmt.Errorf("sshconfig: recursive include cycle at %s", path)
	}
	file, err := openIncludeFile(path)
	if err != nil {
		return fmt.Errorf("sshconfig: read included file %s: %w", path, err)
	}
	defer file.Close()
	if options.CheckPermissions {
		if err := checkIncludePermissions(file, path); err != nil {
			return err
		}
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("sshconfig: read included file %s: %w", path, err)
	}
	doc, err := Parse(data)
	if err != nil {
		return fmt.Errorf("sshconfig: parse included file %s: %w", path, err)
	}
	if _, exists := g.Files[path]; !exists {
		g.Files[path] = &ResolvedFile{Path: path, Document: doc}
	}
	g.Order = append(g.Order, path)
	stack[path] = true
	defer delete(stack, path)

	for _, node := range doc.nodes {
		if node.Directive == nil || node.Directive.KeywordValue != "include" {
			continue
		}
		for _, argument := range node.Directive.Arguments {
			pattern, err := expandIncludePattern(argument.Value, options)
			if err != nil {
				return fmt.Errorf("sshconfig: %s:%d: include %q: %w", path, argument.Position.Line, argument.Value, err)
			}
			if !filepath.IsAbs(pattern) {
				pattern = filepath.Join(options.RelativeBase, pattern)
			}
			matches, err := filepath.Glob(filepath.Clean(pattern))
			if err != nil {
				return fmt.Errorf("sshconfig: %s:%d: invalid include pattern %q: %w", path, argument.Position.Line, pattern, err)
			}
			sort.Strings(matches)
			edge := IncludeEdge{
				FromPath: path,
				NodeID:   node.ID,
				Pattern:  argument.Value,
				Targets:  append([]string(nil), matches...),
			}
			g.Edges = append(g.Edges, edge)
			for _, match := range matches {
				if err := g.resolveFile(match, options, depth+1, stack); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func expandIncludePattern(pattern string, options ResolveOptions) (string, error) {
	var err error
	if options.Environment != nil {
		pattern, err = expandEnvironment(pattern, options.Environment)
		if err != nil {
			return "", err
		}
	}
	pattern, err = expandPercentTokens(pattern, options.Tokens)
	if err != nil {
		return "", err
	}
	if pattern == "~" || strings.HasPrefix(pattern, "~/") || strings.HasPrefix(pattern, `~\`) {
		if options.UserHome == "" {
			return "", fmt.Errorf("cannot expand ~ without UserHome")
		}
		if pattern == "~" {
			pattern = options.UserHome
		} else {
			pattern = filepath.Join(options.UserHome, pattern[2:])
		}
	}
	return pattern, nil
}

func expandEnvironment(value string, environment map[string]string) (string, error) {
	var out strings.Builder
	for index := 0; index < len(value); {
		if value[index] != '$' || index+1 >= len(value) || value[index+1] != '{' {
			out.WriteByte(value[index])
			index++
			continue
		}

		endOffset := strings.IndexByte(value[index+2:], '}')
		if endOffset < 0 {
			return "", fmt.Errorf("unterminated environment variable expansion")
		}
		end := index + 2 + endOffset
		name := value[index+2 : end]
		if !validEnvironmentName(name) {
			return "", fmt.Errorf("invalid environment variable name %q", name)
		}
		replacement, ok := environment[name]
		if !ok {
			return "", fmt.Errorf("environment variable %s is not set", name)
		}
		out.WriteString(replacement)
		index = end + 1
	}
	return out.String(), nil
}

func validEnvironmentName(name string) bool {
	if name == "" || !isEnvironmentNameStart(name[0]) {
		return false
	}
	for index := 1; index < len(name); index++ {
		character := name[index]
		if !isEnvironmentNameStart(character) && (character < '0' || character > '9') {
			return false
		}
	}
	return true
}

func isEnvironmentNameStart(character byte) bool {
	return character == '_' ||
		(character >= 'a' && character <= 'z') ||
		(character >= 'A' && character <= 'Z')
}

func expandPercentTokens(value string, tokens map[byte]string) (string, error) {
	var out strings.Builder
	for i := 0; i < len(value); i++ {
		if value[i] != '%' {
			out.WriteByte(value[i])
			continue
		}
		if i+1 >= len(value) {
			return "", fmt.Errorf("trailing %% token")
		}
		i++
		if value[i] == '%' {
			out.WriteByte('%')
			continue
		}
		replacement, ok := tokens[value[i]]
		if !ok {
			return "", fmt.Errorf("unsupported %%%c token", value[i])
		}
		out.WriteString(replacement)
	}
	return out.String(), nil
}

func checkIncludePermissions(file *os.File, path string) error {
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("sshconfig: inspect included file %s: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("sshconfig: included path %s is not a regular file", path)
	}
	return checkIncludePlatformPermissions(info, path)
}
