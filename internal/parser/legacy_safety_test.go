package parser_test

import (
	"strings"
	"testing"

	Parser "github.com/soulteary/ssh-config/v2/internal/parser"
)

func TestGroupStructuredConfigStrictReturnsDecodeErrors(t *testing.T) {
	t.Parallel()

	if _, err := Parser.GroupJSONConfigStrict("{broken json}"); err == nil {
		t.Fatal("GroupJSONConfigStrict() unexpectedly accepted malformed JSON")
	}
	if _, err := Parser.GroupYAMLConfigStrict("global: ["); err == nil {
		t.Fatal("GroupYAMLConfigStrict() unexpectedly accepted malformed YAML")
	}
}

func TestGroupSSHConfigRejectsLossyLegacyInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "global directive", input: "ServerAliveInterval 30\nHost example\n  User alice\n", want: "global directives"},
		{name: "match", input: "Host first\n  User alice\nMatch host second\n  User bob\n", want: "match cannot be represented"},
		{name: "include", input: "Include config.d/*\n", want: "include cannot be represented"},
		{name: "unknown", input: "Host example\n  SendEnv FOO\n", want: "not supported"},
		{name: "repeated", input: "Host example\n  IdentityFile ~/.ssh/a\n  IdentityFile ~/.ssh/b\n", want: "repeated directive"},
		{name: "duplicate host", input: "Host example\n  User alice\nHost example\n  User bob\n", want: "duplicate Host block"},
		{name: "malformed quote", input: "Host example\n  ProxyCommand \"unterminated\n", want: "unterminated quoted argument"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := Parser.GroupSSHConfig(test.input)
			if err == nil {
				t.Fatal("GroupSSHConfig() unexpectedly succeeded")
			}
			if !strings.Contains(err.Error(), test.want) || !strings.Contains(err.Error(), "use -lossless") {
				t.Fatalf("GroupSSHConfig() error = %q, want %q and lossless hint", err, test.want)
			}
		})
	}
}

func TestGroupSSHConfigAcceptsRepresentableLegacyInput(t *testing.T) {
	t.Parallel()

	configs, err := Parser.GroupSSHConfig("Host example\n  HostName example.com\n  User alice\n  Port 2222\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(configs) != 1 || configs[0].Name != "example" || configs[0].Config["User"] != "alice" {
		t.Fatalf("GroupSSHConfig() = %#v", configs)
	}
}

func TestLegacyConversionPreservesQuotedAndHashArguments(t *testing.T) {
	t.Parallel()

	configs, err := Parser.GroupSSHConfig("Host example\n  IdentityFile \"/tmp/key with space\"\n  ControlPath /tmp/socket#one\n")
	if err != nil {
		t.Fatal(err)
	}
	got := string(Parser.ConvertToSSH(configs))
	for _, want := range []string{`IdentityFile "/tmp/key with space"`, `ControlPath "/tmp/socket#one"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("ConvertToSSH() = %q, missing %q", got, want)
		}
	}

	reparsed, err := Parser.GroupSSHConfig(got)
	if err != nil {
		t.Fatal(err)
	}
	if reparsed[0].Config["IdentityFile"] != "/tmp/key with space" || reparsed[0].Config["ControlPath"] != "/tmp/socket#one" {
		t.Fatalf("round trip = %#v", reparsed[0].Config)
	}
}
