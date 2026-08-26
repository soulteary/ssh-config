package sshconfig

import (
	"bytes"
	"fmt"
	"strings"
)

// ReplaceDirective replaces one directive while retaining the original
// indentation, inline comment, and line ending. The rest of the document is
// not reformatted.
func (d *Document) ReplaceDirective(id NodeID, keyword string, arguments ...string) error {
	if d == nil {
		return fmt.Errorf("sshconfig: nil document")
	}
	if err := ValidateDirectiveInput(keyword, arguments, ""); err != nil {
		return err
	}
	index := d.nodeIndex(id)
	if index < 0 {
		return fmt.Errorf("sshconfig: node %d not found", id)
	}
	node := &d.nodes[index]
	if node.original == nil {
		return fmt.Errorf("sshconfig: node %d is not a source directive", id)
	}
	d.replacement[id] = d.renderReplacement(*node, keyword, arguments)
	node.Kind = NodeDirective
	node.Directive = syntheticDirective(keyword, arguments)
	d.removed[id] = false
	return nil
}

// RemoveNode removes exactly one syntax node. Repeated directives and other
// nodes with the same keyword are unaffected.
func (d *Document) RemoveNode(id NodeID) error {
	if d == nil {
		return fmt.Errorf("sshconfig: nil document")
	}
	if d.nodeIndex(id) < 0 {
		return fmt.Errorf("sshconfig: node %d not found", id)
	}
	d.removed[id] = true
	return nil
}

// InsertDirectiveAfter inserts a new directive after an existing node and
// returns its stable node identifier.
func (d *Document) InsertDirectiveAfter(after NodeID, keyword string, arguments ...string) (NodeID, error) {
	if d == nil {
		return 0, fmt.Errorf("sshconfig: nil document")
	}
	if err := ValidateDirectiveInput(keyword, arguments, ""); err != nil {
		return 0, err
	}
	index := d.nodeIndex(after)
	if index < 0 {
		return 0, fmt.Errorf("sshconfig: node %d not found", after)
	}
	raw := d.renderNewDirective(keyword, arguments)
	if !d.nodeHasLineEnding(d.nodes[index]) {
		raw = append([]byte(d.Newline()), raw...)
	}
	id := d.nextID
	d.nextID++
	node := Node{ID: id, Kind: NodeDirective, Directive: syntheticDirective(keyword, arguments)}
	d.nodes = append(d.nodes, Node{})
	copy(d.nodes[index+2:], d.nodes[index+1:])
	d.nodes[index+1] = node
	d.replacement[id] = raw
	return id, nil
}

// AppendDirective appends a directive to the end of a document.
func (d *Document) AppendDirective(keyword string, arguments ...string) (NodeID, error) {
	if d == nil {
		return 0, fmt.Errorf("sshconfig: nil document")
	}
	if err := ValidateDirectiveInput(keyword, arguments, ""); err != nil {
		return 0, err
	}
	raw := d.renderNewDirective(keyword, arguments)
	for i := len(d.nodes) - 1; i >= 0; i-- {
		if d.removed[d.nodes[i].ID] {
			continue
		}
		if !d.nodeHasLineEnding(d.nodes[i]) {
			raw = append([]byte(d.Newline()), raw...)
		}
		break
	}
	id := d.nextID
	d.nextID++
	d.nodes = append(d.nodes, Node{
		ID:        id,
		Kind:      NodeDirective,
		Directive: syntheticDirective(keyword, arguments),
	})
	d.replacement[id] = raw
	return id, nil
}

func (d *Document) nodeIndex(id NodeID) int {
	for i := range d.nodes {
		if d.nodes[i].ID == id {
			return i
		}
	}
	return -1
}

func (d *Document) nodeHasLineEnding(node Node) bool {
	if raw, ok := d.replacement[node.ID]; ok {
		return bytes.HasSuffix(raw, []byte("\n")) || bytes.HasSuffix(raw, []byte("\r"))
	}
	if node.original != nil {
		return node.original.LineEnding.End > node.original.LineEnding.Start
	}
	return false
}

func (d *Document) renderReplacement(node Node, keyword string, arguments []string) []byte {
	original := node.original
	var out bytes.Buffer
	out.Write(d.source[node.Span.Start:original.Keyword.Span.Start])
	out.WriteString(keyword)
	separator := d.source[original.Separator.Start:original.Separator.End]
	if len(separator) == 0 {
		separator = []byte(" ")
	}
	if len(arguments) > 0 {
		out.Write(separator)
		writeArguments(&out, arguments)
	}
	if original.Comment != nil {
		out.WriteByte(' ')
		out.Write(d.source[original.Comment.Span.Start:original.Comment.Span.End])
	}
	out.Write(d.source[original.LineEnding.Start:original.LineEnding.End])
	return out.Bytes()
}

func (d *Document) renderNewDirective(keyword string, arguments []string) []byte {
	var out bytes.Buffer
	out.WriteString(keyword)
	if len(arguments) > 0 {
		out.WriteByte(' ')
		writeArguments(&out, arguments)
	}
	out.WriteString(d.Newline())
	return out.Bytes()
}

func writeArguments(out *bytes.Buffer, arguments []string) {
	for i, argument := range arguments {
		if i > 0 {
			out.WriteByte(' ')
		}
		out.WriteString(quoteArgument(argument))
	}
}

func quoteArgument(argument string) string {
	if argument != "" && !strings.ContainsAny(argument, " \t\r\n#\\\"'") {
		return argument
	}
	var out strings.Builder
	out.WriteByte('"')
	for _, ch := range argument {
		if ch == '\\' || ch == '"' {
			out.WriteByte('\\')
		}
		out.WriteRune(ch)
	}
	out.WriteByte('"')
	return out.String()
}

func syntheticDirective(keyword string, arguments []string) *Directive {
	parsed := make([]Argument, 0, len(arguments))
	for _, argument := range arguments {
		parsed = append(parsed, Argument{Value: argument})
	}
	return &Directive{
		KeywordValue: strings.ToLower(keyword),
		Arguments:    parsed,
	}
}

// ValidateDirectiveInput rejects fields that could escape their directive line.
func ValidateDirectiveInput(keyword string, arguments []string, comment string) error {
	if keyword == "" {
		return fmt.Errorf("sshconfig: directive keyword is empty")
	}
	for _, ch := range []byte(keyword) {
		if ch <= ' ' || ch == 0x7f || ch == '=' || ch == '#' || ch == '\'' || ch == '"' {
			return fmt.Errorf("sshconfig: directive keyword %q contains an invalid byte", keyword)
		}
	}
	for index, argument := range arguments {
		if strings.ContainsAny(argument, "\x00\r\n") {
			return fmt.Errorf("sshconfig: directive argument %d contains NUL or a line ending", index)
		}
	}
	if strings.ContainsAny(comment, "\x00\r\n") {
		return fmt.Errorf("sshconfig: directive comment contains NUL or a line ending")
	}
	return nil
}
