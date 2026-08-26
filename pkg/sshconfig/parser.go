package sshconfig

import "fmt"

// Parse constructs a lossless syntax document. Unknown or malformed lines are
// retained. Malformed quoting is reported through Document.Diagnostics.
func Parse(input []byte) (*Document, error) {
	doc := &Document{
		source:      append([]byte(nil), input...),
		replacement: make(map[NodeID][]byte),
		removed:     make(map[NodeID]bool),
	}
	line, offset := 1, 0
	for offset < len(doc.source) {
		contentEnd, lineEnd := scanLine(doc.source, offset)
		if doc.lineEnding == "" && lineEnd > contentEnd {
			doc.lineEnding = string(doc.source[contentEnd:lineEnd])
		}
		node := parseLine(doc, NodeID(len(doc.nodes)), line, offset, contentEnd, lineEnd)
		doc.nodes = append(doc.nodes, node)
		offset = lineEnd
		line++
	}
	doc.nextID = NodeID(len(doc.nodes))
	return doc, nil
}

func scanLine(source []byte, start int) (contentEnd, lineEnd int) {
	for i := start; i < len(source); i++ {
		switch source[i] {
		case '\n':
			return i, i + 1
		case '\r':
			if i+1 < len(source) && source[i+1] == '\n' {
				return i, i + 2
			}
			return i, i + 1
		}
	}
	return len(source), len(source)
}

func parseLine(doc *Document, id NodeID, line, start, contentEnd, lineEnd int) Node {
	node := Node{ID: id, Span: Span{Start: start, End: lineEnd}}
	i := skipHorizontalSpace(doc.source, start, contentEnd)
	if i == contentEnd {
		node.Kind = NodeBlank
		return node
	}
	if doc.source[i] == '#' {
		node.Kind = NodeComment
		return node
	}

	keywordStart := i
	for i < contentEnd && !isKeywordDelimiter(doc.source[i]) {
		i++
	}
	if keywordStart == i {
		node.Kind = NodeInvalid
		doc.addDiagnostic(line, keywordStart-start+1, keywordStart, "missing configuration keyword")
		return node
	}

	directive := &Directive{
		Keyword: Token{
			Span:     Span{Start: keywordStart, End: i},
			Position: Position{Offset: keywordStart, Line: line, Column: keywordStart - start + 1},
		},
		KeywordValue: normalizeKeyword(doc.source[keywordStart:i]),
		LineEnding:   Span{Start: contentEnd, End: lineEnd},
	}

	separatorStart := i
	i = skipHorizontalSpace(doc.source, i, contentEnd)
	if i < contentEnd && doc.source[i] == '=' {
		i++
		i = skipHorizontalSpace(doc.source, i, contentEnd)
	}
	directive.Separator = Span{Start: separatorStart, End: i}

	var malformed bool
	directive.Arguments, directive.Comment, malformed = parseArguments(doc, line, start, i, contentEnd)
	node.Kind = NodeDirective
	node.Directive = directive
	node.original = directive
	if malformed {
		node.Kind = NodeInvalid
	}
	return node
}

func parseArguments(doc *Document, line, lineStart, start, end int) ([]Argument, *Token, bool) {
	var arguments []Argument
	i := start
	for {
		for i < end && isArgumentSpace(doc.source[i]) {
			i++
		}
		if i >= end {
			return arguments, nil, false
		}
		if doc.source[i] == '#' {
			comment := &Token{
				Span:     Span{Start: i, End: end},
				Position: Position{Offset: i, Line: line, Column: i - lineStart + 1},
			}
			return arguments, comment, false
		}

		argumentStart := i
		quote := QuoteNone
		quoteByte := byte(0)
		value := make([]byte, 0, end-i)
		for i < end {
			ch := doc.source[i]
			if ch == '\\' {
				if i+1 < end && isEscapable(doc.source[i+1], quoteByte) {
					i++
					value = append(value, doc.source[i])
					i++
					continue
				}
				value = append(value, ch)
				i++
				continue
			}
			if quoteByte == 0 && isArgumentSpace(ch) {
				break
			}
			if quoteByte == 0 && (ch == '\'' || ch == '"') {
				quoteByte = ch
				if quote == QuoteNone {
					if ch == '\'' {
						quote = QuoteSingle
					} else {
						quote = QuoteDouble
					}
				}
				i++
				continue
			}
			if quoteByte != 0 && ch == quoteByte {
				quoteByte = 0
				i++
				continue
			}
			value = append(value, ch)
			i++
		}

		if quoteByte != 0 {
			doc.addDiagnostic(line, argumentStart-lineStart+1, argumentStart, "unterminated quoted argument")
			return arguments, nil, true
		}
		arguments = append(arguments, Argument{
			Token: Token{
				Span:     Span{Start: argumentStart, End: i},
				Position: Position{Offset: argumentStart, Line: line, Column: argumentStart - lineStart + 1},
			},
			Value: string(value),
			Quote: quote,
		})
	}
}

func isArgumentSpace(ch byte) bool {
	return ch == ' ' || ch == '\t'
}

func (d *Document) addDiagnostic(line, column, offset int, message string) {
	d.diagnostics = append(d.diagnostics, Diagnostic{
		Position: Position{Offset: offset, Line: line, Column: column},
		Message:  message,
	})
}

func skipHorizontalSpace(source []byte, start, end int) int {
	for start < end && isHorizontalSpace(source[start]) {
		start++
	}
	return start
}

func isHorizontalSpace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\f'
}

func isKeywordDelimiter(ch byte) bool {
	return isHorizontalSpace(ch) || ch == '=' || ch == '\'' || ch == '"'
}

func isEscapable(ch, quote byte) bool {
	return ch == '\\' || ch == '\'' || ch == '"' || quote == 0 && ch == ' '
}

// ParseFile reads and parses a single configuration file.
func ParseFile(path string, readFile func(string) ([]byte, error)) (*Document, error) {
	if readFile == nil {
		return nil, fmt.Errorf("sshconfig: read function is nil")
	}
	data, err := readFile(path)
	if err != nil {
		return nil, fmt.Errorf("sshconfig: read %s: %w", path, err)
	}
	return Parse(data)
}
