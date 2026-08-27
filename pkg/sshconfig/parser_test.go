package sshconfig

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
)

func TestParsePreservesBytes(t *testing.T) {
	t.Parallel()
	tests := map[string][]byte{
		"empty":              {},
		"blank":              []byte(" \t\r\n\n"),
		"comments":           []byte("# one\n  # two\r\n"),
		"equals":             []byte("Host=example\nSetEnv FOO=bar BAZ=qux\n"),
		"quotes":             []byte("ProxyCommand \"ssh -W %h:%p jump\" # tail\n"),
		"single quotes":      []byte("Match exec 'test -f /tmp/file'\n"),
		"unknown":            []byte("FutureOption   some-value\n"),
		"repeated":           []byte("IdentityFile a\nIdentityFile b\nHost a\nHost a\n"),
		"no trailing line":   []byte("Host final"),
		"invalid quote":      []byte("Host \"unfinished\n"),
		"non utf8 comment":   {0x23, 0x20, 0xff, '\n'},
		"carriage return":    []byte("Host old-mac\rUser root\r"),
		"escaped whitespace": []byte("ProxyCommand ssh\\ jump\n"),
	}
	for name, input := range tests {
		name, input := name, input
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			doc, err := Parse(input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if got := doc.Bytes(); !bytes.Equal(got, input) {
				t.Fatalf("round trip mismatch\n got: %q\nwant: %q", got, input)
			}
		})
	}
}

func TestParseDirectiveSyntax(t *testing.T) {
	t.Parallel()
	input := []byte("  SetEnv = FOO=bar BAZ=qux # comment\r\nHost=one two\n")
	doc, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	nodes := doc.Nodes()
	if len(nodes) != 2 {
		t.Fatalf("node count = %d, want 2", len(nodes))
	}
	first := nodes[0].Directive
	if first == nil || first.KeywordValue != "setenv" {
		t.Fatalf("first directive = %#v", first)
	}
	values := []string{first.Arguments[0].Value, first.Arguments[1].Value}
	if want := []string{"FOO=bar", "BAZ=qux"}; !reflect.DeepEqual(values, want) {
		t.Fatalf("arguments = %v, want %v", values, want)
	}
	if got := string(doc.Raw(first.Comment.Span)); got != "# comment" {
		t.Fatalf("comment = %q", got)
	}
	if got := string(doc.Raw(first.LineEnding)); got != "\r\n" {
		t.Fatalf("line ending = %q", got)
	}
	second := nodes[1].Directive
	if second.KeywordValue != "host" || len(second.Arguments) != 2 {
		t.Fatalf("second directive = %#v", second)
	}
}

func TestNodesReturnsIndependentSyntax(t *testing.T) {
	t.Parallel()
	doc, err := Parse([]byte("Host example # comment\n"))
	if err != nil {
		t.Fatal(err)
	}

	nodes := doc.Nodes()
	nodes[0].Kind = NodeInvalid
	nodes[0].Directive.KeywordValue = "match"
	nodes[0].Directive.Arguments[0].Value = "changed"
	nodes[0].Directive.Arguments = append(nodes[0].Directive.Arguments, Argument{Value: "extra"})
	nodes[0].Directive.Comment.Span = Span{}

	fresh := doc.Nodes()
	if fresh[0].Kind != NodeDirective {
		t.Fatalf("node kind = %v, want NodeDirective", fresh[0].Kind)
	}
	directive := fresh[0].Directive
	if directive.KeywordValue != "host" || len(directive.Arguments) != 1 || directive.Arguments[0].Value != "example" {
		t.Fatalf("directive was mutated through Nodes(): %#v", directive)
	}
	if got := string(doc.Raw(directive.Comment.Span)); got != "# comment" {
		t.Fatalf("comment = %q, want %q", got, "# comment")
	}
	if got := string(doc.Bytes()); got != "Host example # comment\n" {
		t.Fatalf("document bytes = %q", got)
	}
}

func TestNodesPreservesNonNilEmptyArguments(t *testing.T) {
	t.Parallel()
	doc, err := Parse(nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := doc.AppendDirective("Host"); err != nil {
		t.Fatal(err)
	}
	if arguments := doc.Nodes()[0].Directive.Arguments; arguments == nil || len(arguments) != 0 {
		t.Fatalf("Arguments = %#v, want a non-nil empty slice", arguments)
	}
}

func TestMalformedQuoteIsRetainedAndDiagnosed(t *testing.T) {
	t.Parallel()
	input := []byte("Host \"unfinished\n")
	doc, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got := doc.Nodes()[0].Kind; got != NodeInvalid {
		t.Fatalf("kind = %v, want NodeInvalid", got)
	}
	if len(doc.Diagnostics()) != 1 {
		t.Fatalf("diagnostics = %v", doc.Diagnostics())
	}
	if !bytes.Equal(doc.Bytes(), input) {
		t.Fatal("invalid input was not retained")
	}
}

func TestDocumentAccessorsReturnIndependentData(t *testing.T) {
	t.Parallel()

	var nilDocument *Document
	if nilDocument.Bytes() != nil || nilDocument.Nodes() != nil || nilDocument.Diagnostics() != nil || nilDocument.Raw(Span{}) != nil {
		t.Fatal("nil document accessors should return nil")
	}
	if got := nilDocument.Newline(); got != "\n" {
		t.Fatalf("nil document newline = %q, want LF", got)
	}

	document, err := Parse([]byte("Host \"unfinished\n"))
	if err != nil {
		t.Fatal(err)
	}
	bytesCopy := document.Bytes()
	bytesCopy[0] = 'X'
	if got := string(document.Bytes()); got != "Host \"unfinished\n" {
		t.Fatalf("Bytes() exposed document storage: %q", got)
	}

	diagnostics := document.Diagnostics()
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one entry", diagnostics)
	}
	diagnostics[0].Message = "changed"
	if document.Diagnostics()[0].Message == "changed" {
		t.Fatal("Diagnostics() exposed document storage")
	}

	for _, span := range []Span{{Start: -1}, {Start: 2, End: 1}, {End: len(document.Bytes()) + 1}} {
		if raw := document.Raw(span); raw != nil {
			t.Fatalf("Raw(%+v) = %q, want nil", span, raw)
		}
	}
}

func TestParseArgumentsMatchesOpenSSHArgvSplit(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{name: "mid token quotes", input: `SetEnv foo" bar"baz`, want: []string{"foo barbaz"}},
		{name: "adjacent quotes", input: `SetEnv "one"'two'"three"`, want: []string{"onetwothree"}},
		{name: "escaped quotes", input: `SetEnv one\"two one\'three`, want: []string{`one"two`, "one'three"}},
		{name: "escaped space outside quotes", input: `SetEnv one\ two`, want: []string{"one two"}},
		{name: "escaped space inside quotes", input: `SetEnv "one\ two"`, want: []string{`one\ two`}},
		{name: "unrecognized escape", input: `SetEnv one\q`, want: []string{`one\q`}},
		{name: "hash in token", input: `SetEnv value#literal # comment`, want: []string{"value#literal"}},
		{name: "form feed is not argv whitespace", input: "SetEnv one\ftwo", want: []string{"one\ftwo"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			doc, err := Parse([]byte(test.input + "\n"))
			if err != nil {
				t.Fatal(err)
			}
			if diagnostics := doc.Diagnostics(); len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			directive := doc.Nodes()[0].Directive
			got := make([]string, 0, len(directive.Arguments))
			for _, argument := range directive.Arguments {
				got = append(got, argument.Value)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("arguments = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestParseFile(t *testing.T) {
	t.Parallel()
	doc, err := ParseFile("config", func(path string) ([]byte, error) {
		if path != "config" {
			t.Fatalf("path = %q", path)
		}
		return []byte("Host example\n"), nil
	})
	if err != nil || string(doc.Bytes()) != "Host example\n" {
		t.Fatalf("ParseFile() = %q, %v", doc.Bytes(), err)
	}

	wantErr := errors.New("read failed")
	_, err = ParseFile("missing", func(string) ([]byte, error) { return nil, wantErr })
	if !errors.Is(err, wantErr) {
		t.Fatalf("ParseFile() error = %v", err)
	}
	if _, err = ParseFile("missing", nil); err == nil {
		t.Fatal("ParseFile() with nil reader succeeded")
	}
}

func FuzzParsePreservesBytes(f *testing.F) {
	f.Add([]byte("Host example\n    User root\n"))
	f.Add([]byte("Match host=internal # comment\r\n"))
	f.Fuzz(func(t *testing.T, input []byte) {
		doc, err := Parse(input)
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		if got := doc.Bytes(); !bytes.Equal(got, input) {
			t.Fatalf("round trip mismatch\n got: %q\nwant: %q", got, input)
		}
	})
}
