package parser_test

import (
	"strings"
	"testing"

	Parser "github.com/soulteary/ssh-config/v3/internal/parser"
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

func TestGroupStructuredConfigRejectsCrossLineFields(t *testing.T) {
	t.Parallel()

	jsonInput := `[{"Name":"example","Data":{"User":"alice\nHost injected"}}]`
	if _, err := Parser.GroupJSONConfigStrict(jsonInput); err == nil {
		t.Fatal("GroupJSONConfigStrict() accepted a cross-line value")
	}

	yamlInput := "Group test:\n  Hosts:\n    example:\n      config:\n        'User\nProxyCommand': alice\n"
	if _, err := Parser.GroupYAMLConfigStrict(yamlInput); err == nil {
		t.Fatal("GroupYAMLConfigStrict() accepted a cross-line keyword")
	}
}

func TestGroupYAMLConfigAppliesInheritanceToEmptyHost(t *testing.T) {
	t.Parallel()
	input := `default:
  User: alice
Group work:
  Common:
    Port: "22"
  Hosts:
    example: {}
`
	configs, err := Parser.GroupYAMLConfigStrict(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(configs) != 1 {
		t.Fatalf("GroupYAMLConfigStrict() returned %d hosts, want 1", len(configs))
	}
	if configs[0].Name != "example" || configs[0].Config["User"] != "alice" || configs[0].Config["Port"] != "22" {
		t.Fatalf("inherited host = %#v", configs[0])
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

func TestLegacyConversionPreservesMixedQuotedArguments(t *testing.T) {
	t.Parallel()

	input := "Host example\n  IdentityFile /tmp/key\" with \"'space'\n  ProxyCommand env FOO=bar ssh -W %h:%p \"jump host\"\n"
	configs, err := Parser.GroupSSHConfig(input)
	if err != nil {
		t.Fatal(err)
	}
	if got := configs[0].Config["IdentityFile"]; got != "/tmp/key with space" {
		t.Fatalf("IdentityFile = %q", got)
	}
	if got := configs[0].Config["ProxyCommand"]; got != "env FOO=bar ssh -W %h:%p jump host" {
		t.Fatalf("ProxyCommand = %q", got)
	}

	output := string(Parser.ConvertToSSH(configs))
	reparsed, err := Parser.GroupSSHConfig(output)
	if err != nil {
		t.Fatal(err)
	}
	if reparsed[0].Config["IdentityFile"] != configs[0].Config["IdentityFile"] ||
		reparsed[0].Config["ProxyCommand"] != configs[0].Config["ProxyCommand"] {
		t.Fatalf("round trip changed arguments:\n%s", output)
	}
}

func TestLegacyRoundTripPreservesHostOrder(t *testing.T) {
	t.Parallel()

	input := "Host special.example.com\n  User special\nHost *.example.com\n  User wildcard\n"
	configs, err := Parser.GroupSSHConfig(input)
	if err != nil {
		t.Fatal(err)
	}
	yamlData := Parser.ConvertToYAML(configs)
	fromYAML, err := Parser.GroupYAMLConfigStrict(string(yamlData))
	if err != nil {
		t.Fatal(err)
	}
	output := string(Parser.ConvertToSSH(fromYAML))
	special := strings.Index(output, "Host special.example.com")
	wildcard := strings.Index(output, "Host *.example.com")
	if special < 0 || wildcard < 0 || special > wildcard {
		t.Fatalf("host order changed:\n%s", output)
	}
}

func TestLegacyYAMLPreservesHostsMappingOrder(t *testing.T) {
	t.Parallel()

	input := "Group work:\n  Hosts:\n    z-host:\n      config:\n        User: z\n    a-host:\n      config:\n        User: a\n"
	configs, err := Parser.GroupYAMLConfigStrict(input)
	if err != nil {
		t.Fatal(err)
	}
	output := string(Parser.ConvertToSSH(configs))
	if strings.Index(output, "Host z-host") > strings.Index(output, "Host a-host") {
		t.Fatalf("host order changed:\n%s", output)
	}
}

func TestLegacyYAMLPreservesMergedHostsOrder(t *testing.T) {
	t.Parallel()

	input := `Group template: &group
  Hosts:
    z-host:
      config:
        User: z
    a-host:
      config:
        User: a
Group inherited:
  <<: *group
`
	configs, err := Parser.GroupYAMLConfigStrict(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(configs) != 4 || configs[2].Name != "z-host" || configs[3].Name != "a-host" {
		t.Fatalf("merged host order = %#v", configs)
	}
}
