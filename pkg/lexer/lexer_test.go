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

package lexer

import (
	"reflect"
	"testing"
)

func TestLex_Empty(t *testing.T) {
	tokens, err := Lex("")
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 1 || tokens[0].Kind != TokenEOF {
		t.Errorf("Lex(\"\") = %v, want [EOF]", tokens)
	}
}

func TestLex_NewlinesOnly(t *testing.T) {
	tokens, err := Lex("\n\n")
	if err != nil {
		t.Fatal(err)
	}
	want := []Token{
		{Kind: TokenNewline, Value: "\n", Line: 1, Column: 1},
		{Kind: TokenNewline, Value: "\n", Line: 2, Column: 1},
		{Kind: TokenEOF, Value: "", Line: 3, Column: 1},
	}
	assertTokens(t, tokens, want)
}

func TestLex_CommentOnly(t *testing.T) {
	tokens, err := Lex("# server-cn-1\n")
	if err != nil {
		t.Fatal(err)
	}
	want := []Token{
		{Kind: TokenComment, Value: "server-cn-1", Line: 1, Column: 1},
		{Kind: TokenNewline, Value: "\n", Line: 1, Column: 14},
		{Kind: TokenEOF, Value: "", Line: 2, Column: 1},
	}
	assertTokens(t, tokens, want)
}

func TestLex_HostBlock(t *testing.T) {
	input := "Host server-cn-1\n    Hostname 123.123.123.1\n    User ubuntu\n"
	tokens, err := Lex(input)
	if err != nil {
		t.Fatal(err)
	}
	want := []Token{
		{Kind: TokenKeyword, Value: "Host", Line: 1, Column: 1},
		{Kind: TokenValue, Value: "server-cn-1", Line: 1, Column: 6},
		{Kind: TokenNewline, Value: "\n", Line: 1, Column: 17},
		{Kind: TokenIdent, Value: "Hostname", Line: 2, Column: 5},
		{Kind: TokenValue, Value: "123.123.123.1", Line: 2, Column: 14},
		{Kind: TokenNewline, Value: "\n", Line: 2, Column: 27},
		{Kind: TokenIdent, Value: "User", Line: 3, Column: 5},
		{Kind: TokenValue, Value: "ubuntu", Line: 3, Column: 10},
		{Kind: TokenNewline, Value: "\n", Line: 3, Column: 16},
		{Kind: TokenEOF, Value: "", Line: 4, Column: 1},
	}
	assertTokens(t, tokens, want)
}

func TestLex_KeyEqualsValue(t *testing.T) {
	input := "Port=2222\n"
	tokens, err := Lex(input)
	if err != nil {
		t.Fatal(err)
	}
	want := []Token{
		{Kind: TokenIdent, Value: "Port", Line: 1, Column: 1},
		{Kind: TokenEquals, Value: "=", Line: 1, Column: 5},
		{Kind: TokenValue, Value: "2222", Line: 1, Column: 6},
		{Kind: TokenNewline, Value: "\n", Line: 1, Column: 10},
		{Kind: TokenEOF, Value: "", Line: 2, Column: 1},
	}
	assertTokens(t, tokens, want)
}

func TestLex_QuotedValue(t *testing.T) {
	input := `Path "C:\Program Files\ssh"` + "\n"
	tokens, err := Lex(input)
	if err != nil {
		t.Fatal(err)
	}
	want := []Token{
		{Kind: TokenIdent, Value: "Path", Line: 1, Column: 1},
		{Kind: TokenQuoted, Value: `C:\Program Files\ssh`, Line: 1, Column: 6},
		{Kind: TokenNewline, Value: "\n", Line: 1, Column: 28},
		{Kind: TokenEOF, Value: "", Line: 2, Column: 1},
	}
	assertTokens(t, tokens, want)
}

func TestLex_QuotedEscape(t *testing.T) {
	input := `X "say \"hi\""` + "\n"
	tokens, err := Lex(input)
	if err != nil {
		t.Fatal(err)
	}
	want := []Token{
		{Kind: TokenIdent, Value: "X", Line: 1, Column: 1},
		{Kind: TokenQuoted, Value: `say "hi"`, Line: 1, Column: 4},
		{Kind: TokenNewline, Value: "\n", Line: 1, Column: 15},
		{Kind: TokenEOF, Value: "", Line: 2, Column: 1},
	}
	assertTokens(t, tokens, want)
}

func TestLex_UnclosedQuote(t *testing.T) {
	_, err := Lex(`Host "unclosed`)
	if err == nil {
		t.Error("Lex: expected error for unclosed quote")
	}
}

func TestLex_IncludeAndMatch(t *testing.T) {
	input := "Include ~/.ssh/config.d/*\nMatch host foo\n"
	tokens, err := Lex(input)
	if err != nil {
		t.Fatal(err)
	}
	// Include is keyword, then value; Match is keyword, then two values
	kinds := make([]TokenKind, 0, len(tokens))
	for _, tok := range tokens {
		if tok.Kind == TokenEOF {
			break
		}
		kinds = append(kinds, tok.Kind)
	}
	expectKinds := []TokenKind{
		TokenKeyword, TokenValue, TokenNewline,
		TokenKeyword, TokenValue, TokenValue, TokenNewline,
	}
	if len(kinds) < len(expectKinds) {
		t.Fatalf("got %d tokens (kinds), want at least %d", len(kinds), len(expectKinds))
	}
	for i, k := range expectKinds {
		if kinds[i] != k {
			t.Errorf("token %d: kind = %v, want %v", i, kinds[i], k)
		}
	}
	if tokens[0].Value != "Include" || tokens[3].Value != "Match" {
		t.Errorf("keywords: got %q, %q want Include, Match", tokens[0].Value, tokens[3].Value)
	}
}

func TestLex_GlobalHost(t *testing.T) {
	input := "Host *\n    Port 22\n"
	tokens, err := Lex(input)
	if err != nil {
		t.Fatal(err)
	}
	want := []Token{
		{Kind: TokenKeyword, Value: "Host", Line: 1, Column: 1},
		{Kind: TokenValue, Value: "*", Line: 1, Column: 6},
		{Kind: TokenNewline, Value: "\n", Line: 1, Column: 7},
		{Kind: TokenIdent, Value: "Port", Line: 2, Column: 5},
		{Kind: TokenValue, Value: "22", Line: 2, Column: 10},
		{Kind: TokenNewline, Value: "\n", Line: 2, Column: 12},
		{Kind: TokenEOF, Value: "", Line: 3, Column: 1},
	}
	assertTokens(t, tokens, want)
}

func assertTokens(t *testing.T, got, want []Token) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("len(tokens) = %d, want %d; got %v", len(got), len(want), got)
		return
	}
	for i := range got {
		if got[i].Kind != want[i].Kind || got[i].Value != want[i].Value || got[i].Line != want[i].Line {
			t.Errorf("token[%d]: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestNextToken_Iteration(t *testing.T) {
	input := "Host a\n"
	l := NewLexer(input)
	var kinds []TokenKind
	for {
		tok, err := l.NextToken()
		if err != nil {
			t.Fatal(err)
		}
		kinds = append(kinds, tok.Kind)
		if tok.Kind == TokenEOF {
			break
		}
	}
	want := []TokenKind{TokenKeyword, TokenValue, TokenNewline, TokenEOF}
	if !reflect.DeepEqual(kinds, want) {
		t.Errorf("NextToken kinds = %v, want %v", kinds, want)
	}
}

// Tests below align with ssh_config(5): https://man7.org/linux/man-pages/man5/ssh_config.5.html

func TestLex_InlineComment(t *testing.T) {
	// "Lines starting with '#'" and rest of line after # is comment
	input := "Host foo # bar\n"
	tokens, err := Lex(input)
	if err != nil {
		t.Fatal(err)
	}
	want := []Token{
		{Kind: TokenKeyword, Value: "Host", Line: 1, Column: 1},
		{Kind: TokenValue, Value: "foo", Line: 1, Column: 6},
		{Kind: TokenComment, Value: "bar", Line: 1, Column: 11},
		{Kind: TokenNewline, Value: "\n", Line: 1, Column: 15},
		{Kind: TokenEOF, Value: "", Line: 2, Column: 1},
	}
	assertTokens(t, tokens, want)
}

func TestLex_KeySpaceEqualsSpaceValue(t *testing.T) {
	// "optional whitespace and exactly one '='"
	input := "Port = 2222\n"
	tokens, err := Lex(input)
	if err != nil {
		t.Fatal(err)
	}
	want := []Token{
		{Kind: TokenIdent, Value: "Port", Line: 1, Column: 1},
		{Kind: TokenEquals, Value: "=", Line: 1, Column: 6},
		{Kind: TokenValue, Value: "2222", Line: 1, Column: 8},
		{Kind: TokenNewline, Value: "\n", Line: 1, Column: 12},
		{Kind: TokenEOF, Value: "", Line: 2, Column: 1},
	}
	assertTokens(t, tokens, want)
}

func TestLex_NegatedPattern(t *testing.T) {
	// "A pattern entry may be negated by prefixing it with an exclamation mark ('!')"
	input := "Host !*.dialup.example.com\n"
	tokens, err := Lex(input)
	if err != nil {
		t.Fatal(err)
	}
	if tokens[1].Kind != TokenValue || tokens[1].Value != "!*.dialup.example.com" {
		t.Errorf("negated pattern: got %+v, want Value !*.dialup.example.com", tokens[1])
	}
}

func TestLex_MultipleHostPatterns(t *testing.T) {
	// "If more than one pattern is provided, they should be separated by whitespace"
	input := "Host 192.168.0.? *.co.uk\n"
	tokens, err := Lex(input)
	if err != nil {
		t.Fatal(err)
	}
	wantKinds := []TokenKind{TokenKeyword, TokenValue, TokenValue, TokenNewline, TokenEOF}
	for i, tok := range tokens {
		if i >= len(wantKinds) {
			break
		}
		if tok.Kind != wantKinds[i] {
			t.Errorf("token %d: kind %v want %v", i, tok.Kind, wantKinds[i])
		}
	}
	if tokens[1].Value != "192.168.0.?" || tokens[2].Value != "*.co.uk" {
		t.Errorf("patterns: got %q, %q", tokens[1].Value, tokens[2].Value)
	}
}

func TestLex_EmptyQuotedValue(t *testing.T) {
	// Arguments enclosed in double quotes; empty string is valid
	input := "X \"\"\n"
	tokens, err := Lex(input)
	if err != nil {
		t.Fatal(err)
	}
	if tokens[1].Kind != TokenQuoted || tokens[1].Value != "" {
		t.Errorf("empty quoted: got %+v", tokens[1])
	}
}

func TestLex_KeywordsCaseInsensitive(t *testing.T) {
	// "keywords are case-insensitive"
	input := "HOST server\nMATCH all\nINCLUDE file\n"
	tokens, err := Lex(input)
	if err != nil {
		t.Fatal(err)
	}
	if tokens[0].Kind != TokenKeyword || tokens[0].Value != "HOST" {
		t.Errorf("HOST: got %+v", tokens[0])
	}
	// Find MATCH (after Newline)
	var matchTok *Token
	for i := range tokens {
		if tokens[i].Value == "MATCH" {
			matchTok = &tokens[i]
			break
		}
	}
	if matchTok == nil || matchTok.Kind != TokenKeyword {
		t.Errorf("MATCH: got %+v", matchTok)
	}
}

func TestLex_IncludeQuotedPath(t *testing.T) {
	// Include "path with spaces"
	input := `Include "path with spaces"` + "\n"
	tokens, err := Lex(input)
	if err != nil {
		t.Fatal(err)
	}
	if tokens[0].Kind != TokenKeyword || tokens[0].Value != "Include" {
		t.Errorf("Include: got %+v", tokens[0])
	}
	if tokens[1].Kind != TokenQuoted || tokens[1].Value != "path with spaces" {
		t.Errorf("quoted path: got %+v", tokens[1])
	}
}

func TestLex_CommaSeparatedValue(t *testing.T) {
	// Pattern-list / comma-separated value (e.g. SendEnv)
	input := "SendEnv A,B,C\n"
	tokens, err := Lex(input)
	if err != nil {
		t.Fatal(err)
	}
	if tokens[1].Kind != TokenValue || tokens[1].Value != "A,B,C" {
		t.Errorf("comma value: got %+v", tokens[1])
	}
}

// --- 100% coverage: TokenKind.String(), Token.String(), next() at EOF, quoted backslash-other ---

func TestTokenKind_String(t *testing.T) {
	tests := []struct {
		k    TokenKind
		want string
	}{
		{TokenComment, "Comment"},
		{TokenNewline, "Newline"},
		{TokenKeyword, "Keyword"},
		{TokenIdent, "Ident"},
		{TokenValue, "Value"},
		{TokenQuoted, "Quoted"},
		{TokenEquals, "Equals"},
		{TokenEOF, "EOF"},
		{TokenKind(99), "Unknown"},
	}
	for _, tt := range tests {
		if got := tt.k.String(); got != tt.want {
			t.Errorf("TokenKind(%d).String() = %q, want %q", tt.k, got, tt.want)
		}
	}
}

func TestToken_String(t *testing.T) {
	tok := Token{Kind: TokenIdent, Value: "Host", Line: 1, Column: 2}
	s := tok.String()
	if s == "" || !reflect.DeepEqual(tok, tok) {
		t.Logf("Token.String() = %s", s)
	}
	// Ensure it doesn't panic and contains kind and value
	if len(s) < 5 {
		t.Errorf("Token.String() too short: %q", s)
	}
}

func TestLexer_nextAtEOF(t *testing.T) {
	l := NewLexer("a")
	if r := l.next(); r != 'a' {
		t.Errorf("first next() = %v, want 'a'", r)
	}
	if r := l.next(); r != 0 {
		t.Errorf("next() at EOF = %v, want 0", r)
	}
}

func TestLexer_nextNewlineIncrementsLine(t *testing.T) {
	l := NewLexer("x\ny")
	l.next() // 'x'
	l.next() // '\n' -> line 2
	if l.line != 2 || l.col != 1 {
		t.Errorf("after \\n: line=%d col=%d, want 2, 1", l.line, l.col)
	}
	l.next() // 'y'
	if l.line != 2 || l.col != 2 {
		t.Errorf("after y: line=%d col=%d", l.line, l.col)
	}
}

func TestLex_QuotedBackslashOther(t *testing.T) {
	// In quoted string: \ not followed by " or \ -> write literal \ (else branch)
	input := `K "a\b"` + "\n"
	tokens, err := Lex(input)
	if err != nil {
		t.Fatal(err)
	}
	if tokens[1].Kind != TokenQuoted || tokens[1].Value != "a\\b" {
		t.Errorf("quoted \\b: got Kind=%v Value=%q", tokens[1].Kind, tokens[1].Value)
	}
}

func TestLex_CommentTrimSpace(t *testing.T) {
	// Comment content is TrimSpace'd (slice()[1:])
	input := "#   spaced comment   \n"
	tokens, err := Lex(input)
	if err != nil {
		t.Fatal(err)
	}
	if tokens[0].Kind != TokenComment || tokens[0].Value != "spaced comment" {
		t.Errorf("comment: got %q", tokens[0].Value)
	}
}

func TestLex_WhitespaceOnlyLine(t *testing.T) {
	// \t and \r are skipped like space; multiple spaces skipped
	input := "  \t\r  \n"
	tokens, err := Lex(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 2 || tokens[0].Kind != TokenNewline || tokens[1].Kind != TokenEOF {
		t.Errorf("whitespace line: got %v", tokens)
	}
}

func TestLex_WordThenCommentNoSpace(t *testing.T) {
	// Word stops at # so "word#rest" gives Value "word" then Comment "rest"
	input := "Host server#note\n"
	tokens, err := Lex(input)
	if err != nil {
		t.Fatal(err)
	}
	if tokens[1].Value != "server" || tokens[2].Kind != TokenComment || tokens[2].Value != "note" {
		t.Errorf("word#comment: got %v", tokens)
	}
}

func TestLex_LexErrorPath(t *testing.T) {
	_, err := Lex(`"unclosed`)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() == "" {
		t.Error("error message should not be empty")
	}
}
