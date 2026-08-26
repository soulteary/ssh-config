package sshconfig

import (
	"bytes"
	"testing"
)

func TestLookupKeyword(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		status KeywordStatus
	}{
		{"ProxyJump", KeywordSupported},
		{"MATCH", KeywordSupported},
		{"Protocol", KeywordIgnored},
		{"Cipher", KeywordDeprecated},
		{"KerberosAuthentication", KeywordUnsupported},
		{"PKCS11Provider", KeywordPlatformDependent},
	}
	for _, test := range tests {
		info, ok := LookupKeyword(test.name)
		if !ok || info.Status != test.status {
			t.Errorf("LookupKeyword(%q) = %#v, %v", test.name, info, ok)
		}
	}
	if _, ok := LookupKeyword("FutureOption"); ok {
		t.Fatal("future option unexpectedly registered")
	}
}

func TestValidateReportsWithoutChangingSource(t *testing.T) {
	t.Parallel()
	input := []byte("Host example\nFutureOption value\nCipher old\nProtocol 2\nKerberosAuthentication yes\n")
	doc, _ := Parse(input)
	issues := doc.Validate(ValidateOptions{UnknownDirective: UnknownDirectiveError})
	wantCodes := map[string]bool{
		"unknown-directive":     false,
		"deprecated-directive":  false,
		"ignored-directive":     false,
		"unsupported-directive": false,
	}
	for _, issue := range issues {
		if _, ok := wantCodes[issue.Code]; ok {
			wantCodes[issue.Code] = true
		}
	}
	for code, found := range wantCodes {
		if !found {
			t.Errorf("missing issue %q: %#v", code, issues)
		}
	}
	got, err := doc.MarshalPreserve()
	if err != nil || !bytes.Equal(got, input) {
		t.Fatalf("validation changed source: %q, %v", got, err)
	}
}

func TestValidateSyntaxAndArguments(t *testing.T) {
	t.Parallel()
	doc, _ := Parse([]byte("Host\nUser \"unfinished\n"))
	issues := doc.Validate(ValidateOptions{})
	var syntax, missing bool
	for _, issue := range issues {
		syntax = syntax || issue.Code == "invalid-syntax"
		missing = missing || issue.Code == "missing-argument"
	}
	if !syntax || !missing {
		t.Fatalf("issues = %#v", issues)
	}
}

func TestUnknownDirectivePolicies(t *testing.T) {
	t.Parallel()
	doc, _ := Parse([]byte("FutureOption value\n"))
	if got := doc.Validate(ValidateOptions{UnknownDirective: UnknownDirectiveIgnore}); len(got) != 0 {
		t.Fatalf("ignore issues = %#v", got)
	}
	warn := doc.Validate(ValidateOptions{UnknownDirective: UnknownDirectiveWarn})
	if len(warn) != 1 || warn[0].Severity != SeverityWarning {
		t.Fatalf("warn issues = %#v", warn)
	}
	fail := doc.Validate(ValidateOptions{UnknownDirective: UnknownDirectiveError})
	if len(fail) != 1 || fail[0].Severity != SeverityError {
		t.Fatalf("error issues = %#v", fail)
	}
}
