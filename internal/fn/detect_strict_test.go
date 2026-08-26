package fn_test

import (
	"strings"
	"testing"

	Fn "github.com/soulteary/ssh-config/v2/internal/fn"
)

func TestDetectStringTypeStrictRejectsMalformedStructuredInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "json object", input: "{broken json}\n", want: "invalid JSON input"},
		{name: "json array", input: "[{}", want: "invalid JSON input"},
		{name: "legacy yaml", input: "global: [\n", want: "invalid YAML input"},
		{name: "legacy group yaml", input: "Group work:\n  Hosts: [\n", want: "invalid YAML input"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := Fn.DetectStringTypeStrict(test.input)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("DetectStringTypeStrict() error = %v, want substring %q", err, test.want)
			}
		})
	}
}

func TestDetectStringTypeStrictKeepsSSHText(t *testing.T) {
	t.Parallel()

	got, err := Fn.DetectStringTypeStrict("Host example\n  User alice\n")
	if err != nil {
		t.Fatal(err)
	}
	if got != "TEXT" {
		t.Fatalf("DetectStringTypeStrict() = %q, want TEXT", got)
	}
}
