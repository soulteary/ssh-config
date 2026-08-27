package sshconfig

import (
	"strings"
	"testing"
)

func TestParseReaderLimitsInput(t *testing.T) {
	t.Parallel()
	input := "Host example\n"

	if _, err := ParseReader(strings.NewReader(input), ParseOptions{MaxBytes: int64(len(input))}); err != nil {
		t.Fatalf("exact limit: %v", err)
	}
	if _, err := ParseReader(strings.NewReader(input), ParseOptions{MaxBytes: int64(len(input) - 1)}); err == nil || !strings.Contains(err.Error(), "exceeds maximum size") {
		t.Fatalf("over limit error = %v", err)
	}
	if _, err := ParseReader(strings.NewReader(input), ParseOptions{}); err != nil {
		t.Fatalf("unlimited input: %v", err)
	}
	if _, err := ParseReader(nil, ParseOptions{}); err == nil || !strings.Contains(err.Error(), "nil") {
		t.Fatalf("nil reader error = %v", err)
	}
	if _, err := ParseReader(strings.NewReader(input), ParseOptions{MaxBytes: -1}); err == nil || !strings.Contains(err.Error(), "negative") {
		t.Fatalf("negative limit error = %v", err)
	}
}
