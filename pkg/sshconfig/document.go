package sshconfig

import (
	"bytes"
	"strings"
)

// NodeID identifies a node within a Document.
type NodeID int

// Span is a half-open byte range in the original document.
type Span struct {
	Start int
	End   int
}

// Position identifies a byte position in the original document.
type Position struct {
	Offset int
	Line   int
	Column int
}

// NodeKind describes a top-level syntax node.
type NodeKind uint8

const (
	NodeBlank NodeKind = iota
	NodeComment
	NodeDirective
	NodeInvalid
)

// QuoteStyle records how an argument was quoted in the source.
type QuoteStyle uint8

const (
	QuoteNone QuoteStyle = iota
	QuoteSingle
	QuoteDouble
)

// Token is a source-backed syntax token.
type Token struct {
	Span     Span
	Position Position
}

// Argument is one parsed directive argument. Value is its decoded value;
// Span and Quote retain its original representation.
type Argument struct {
	Token
	Value string
	Quote QuoteStyle
}

// Directive is a parsed directive line. KeywordValue is lower-cased for
// comparisons while Keyword retains its original spelling.
type Directive struct {
	Keyword      Token
	KeywordValue string
	Separator    Span
	Arguments    []Argument
	Comment      *Token
	LineEnding   Span
}

// Node is one physical line in a configuration file. Its span includes the
// original line ending, when present.
type Node struct {
	ID        NodeID
	Kind      NodeKind
	Span      Span
	Directive *Directive
	original  *Directive
}

// Diagnostic records a recoverable syntax problem. Lossless parsing keeps
// the affected bytes and returns them as an invalid node.
type Diagnostic struct {
	Position Position
	Message  string
}

// Document is a lossless, source-backed syntax tree.
type Document struct {
	source      []byte
	nodes       []Node
	diagnostics []Diagnostic
	lineEnding  string
	replacement map[NodeID][]byte
	removed     map[NodeID]bool
	nextID      NodeID
}

// Bytes returns an independent copy of the original document bytes.
func (d *Document) Bytes() []byte {
	if d == nil {
		return nil
	}
	return bytes.Clone(d.source)
}

// Nodes returns an independent copy of the document nodes.
func (d *Document) Nodes() []Node {
	if d == nil {
		return nil
	}
	nodes := make([]Node, len(d.nodes))
	for index, node := range d.nodes {
		nodes[index] = cloneNode(node)
	}
	return nodes
}

func cloneNode(node Node) Node {
	node.Directive = cloneDirective(node.Directive)
	node.original = cloneDirective(node.original)
	return node
}

func cloneDirective(directive *Directive) *Directive {
	if directive == nil {
		return nil
	}
	clone := *directive
	if directive.Arguments != nil {
		clone.Arguments = append([]Argument{}, directive.Arguments...)
	}
	if directive.Comment != nil {
		comment := *directive.Comment
		clone.Comment = &comment
	}
	return &clone
}

// Diagnostics returns a copy of recoverable parse diagnostics.
func (d *Document) Diagnostics() []Diagnostic {
	if d == nil {
		return nil
	}
	return append([]Diagnostic(nil), d.diagnostics...)
}

// Raw returns an independent copy of the bytes covered by span.
func (d *Document) Raw(span Span) []byte {
	if d == nil || span.Start < 0 || span.End < span.Start || span.End > len(d.source) {
		return nil
	}
	return bytes.Clone(d.source[span.Start:span.End])
}

// Newline returns the first line-ending style found in the source. It
// defaults to LF for an empty or single-line document.
func (d *Document) Newline() string {
	if d == nil || d.lineEnding == "" {
		return "\n"
	}
	return d.lineEnding
}

func normalizeKeyword(value []byte) string {
	return strings.ToLower(string(value))
}
