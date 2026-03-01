/**
 * Copyright 2026 Su Yang (soulteary)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package lexer tokenizes SSH client configuration files.
// Format follows OpenSSH ssh_config(5): https://man7.org/linux/man-pages/man5/ssh_config.5.html
// - Comments: # to end of line; empty lines ignored.
// - Keywords (Host, Match, Include) are case-insensitive; arguments are case-sensitive.
// - Options separated by whitespace or optional whitespace and exactly one '='.
// - Arguments containing spaces may be enclosed in double quotes; \" and \\ are escaped.
package lexer

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// TokenKind is the type of a lexer token.
type TokenKind int

const (
	TokenComment TokenKind = iota
	TokenNewline
	TokenKeyword
	TokenIdent
	TokenValue
	TokenQuoted
	TokenEquals
	TokenEOF
)

func (k TokenKind) String() string {
	switch k {
	case TokenComment:
		return "Comment"
	case TokenNewline:
		return "Newline"
	case TokenKeyword:
		return "Keyword"
	case TokenIdent:
		return "Ident"
	case TokenValue:
		return "Value"
	case TokenQuoted:
		return "Quoted"
	case TokenEquals:
		return "Equals"
	case TokenEOF:
		return "EOF"
	default:
		return "Unknown"
	}
}

// Token represents a single lexer token with optional position.
type Token struct {
	Kind   TokenKind
	Value  string
	Line   int
	Column int
}

func (t Token) String() string {
	return fmt.Sprintf("%s(%q)@%d:%d", t.Kind, t.Value, t.Line, t.Column)
}

// Block keywords at line start (case-insensitive).
var blockKeywords = map[string]struct{}{
	"host": {}, "match": {}, "include": {},
}

// Lexer scans SSH config text and produces tokens.
type Lexer struct {
	input       string
	pos         int
	line        int
	col         int
	start       int
	startLine   int
	startCol    int
	atLineStart bool
}

// NewLexer returns a lexer for the given input.
func NewLexer(input string) *Lexer {
	return &Lexer{
		input:       input,
		line:        1,
		col:         1,
		atLineStart: true,
	}
}

func (l *Lexer) peek() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.pos:])
	return r
}

func (l *Lexer) next() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += w
	l.col += 1
	if r == '\n' {
		l.line++
		l.col = 1
	}
	return r
}

func (l *Lexer) emit(kind TokenKind, value string) Token {
	return Token{Kind: kind, Value: value, Line: l.startLine, Column: l.startCol}
}

func (l *Lexer) slice() string {
	return l.input[l.start:l.pos]
}

// NextToken returns the next token and advances the lexer. Returns TokenEOF and nil error when done.
func (l *Lexer) NextToken() (Token, error) {
	for {
		l.start = l.pos
		l.startLine, l.startCol = l.line, l.col
		r := l.peek()
		if r == 0 {
			return l.emit(TokenEOF, ""), nil
		}

		switch {
		case r == '\n':
			l.next()
			l.atLineStart = true
			return l.emit(TokenNewline, "\n"), nil

		case r == ' ' || r == '\t' || r == '\r':
			for l.peek() == ' ' || l.peek() == '\t' || l.peek() == '\r' {
				l.next()
			}
			continue

		case r == '#':
			l.next()
			for l.peek() != 0 && l.peek() != '\n' {
				l.next()
			}
			comment := strings.TrimSpace(l.slice()[1:])
			return l.emit(TokenComment, comment), nil

		case r == '"':
			l.next()
			var b strings.Builder
			for {
				r = l.peek()
				if r == 0 {
					return Token{}, fmt.Errorf("unclosed quoted string at line %d column %d", l.line, l.col)
				}
				if r == '"' {
					l.next()
					break
				}
				if r == '\\' {
					l.next()
					r = l.peek()
					if r == '"' || r == '\\' {
						l.next()
						b.WriteRune(r)
					} else {
						b.WriteRune('\\')
					}
					continue
				}
				l.next()
				b.WriteRune(r)
			}
			return l.emit(TokenQuoted, b.String()), nil

		case r == '=':
			l.next()
			l.atLineStart = false
			return l.emit(TokenEquals, "="), nil

		default:
			for l.peek() != 0 && l.peek() != '\n' && l.peek() != ' ' && l.peek() != '\t' && l.peek() != '\r' && l.peek() != '#' && l.peek() != '=' {
				l.next()
			}
			word := l.slice()
			// In valid UTF-8 we always advance at least one rune in the loop above, so word != "".

			atStart := l.atLineStart
			l.atLineStart = false

			if atStart {
				lower := strings.ToLower(word)
				if _, ok := blockKeywords[lower]; ok {
					return l.emit(TokenKeyword, word), nil
				}
				return l.emit(TokenIdent, word), nil
			}
			return l.emit(TokenValue, word), nil
		}
	}
}

// Lex returns all tokens from input. Stops on first error (e.g. unclosed quote).
func Lex(input string) ([]Token, error) {
	l := NewLexer(input)
	var tokens []Token
	for {
		tok, err := l.NextToken()
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, tok)
		if tok.Kind == TokenEOF {
			break
		}
	}
	return tokens, nil
}
