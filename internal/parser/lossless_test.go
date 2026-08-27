package parser_test

import (
	"bytes"
	"strings"
	"testing"

	Cmd "github.com/soulteary/ssh-config/v2/cmd"
	Parser "github.com/soulteary/ssh-config/v2/internal/parser"
)

func TestProcessLosslessSSHThroughStructuredFormats(t *testing.T) {
	t.Parallel()
	input := "# keep\r\nHost=example\r\n\tIdentityFile first\r\n\tIdentityFile second # backup\r\n"

	yamlData, err := Parser.Process("TEXT", input, Cmd.Args{ToYAML: true})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(yamlData), "schemaVersion: 3") || !strings.Contains(string(yamlData), "rawBase64:") {
		t.Fatalf("lossless YAML does not contain the v3 envelope:\n%s", yamlData)
	}
	yamlBack, err := Parser.Process("YAML", string(yamlData), Cmd.Args{ToSSH: true})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(yamlBack, []byte(input)) {
		t.Fatalf("YAML round trip mismatch\n got: %q\nwant: %q", yamlBack, input)
	}

	jsonData, err := Parser.Process("TEXT", input, Cmd.Args{ToJSON: true})
	if err != nil {
		t.Fatal(err)
	}
	jsonBack, err := Parser.Process("YAML", string(jsonData), Cmd.Args{ToSSH: true})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(jsonBack, []byte(input)) {
		t.Fatalf("JSON round trip mismatch\n got: %q\nwant: %q", jsonBack, input)
	}
}

func TestProcessLosslessMigratesLegacyJSON(t *testing.T) {
	t.Parallel()
	input := `[{"Name":"example","Data":{"User":"root"}}]`
	got, err := Parser.Process("JSON", input, Cmd.Args{ToSSH: true})
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "Host example\n    User root\n" {
		t.Fatalf("migrated output = %q", got)
	}
}

func TestProcessLosslessRejectsInvalidStructuredInput(t *testing.T) {
	t.Parallel()
	_, err := Parser.Process("YAML", "schemaVersion: 9\ndocuments: []\n", Cmd.Args{ToSSH: true})
	if err == nil {
		t.Fatal("invalid structured input was accepted")
	}
}

func TestProcessLosslessDetectsReorderedV3YAML(t *testing.T) {
	t.Parallel()
	input := `documents:
- path: ""
  nodes:
  - type: directive
    directive:
      keyword: Host
      arguments:
      - example
schemaVersion: 3
`
	got, err := Parser.Process("TEXT", input, Cmd.Args{ToSSH: true})
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "Host example\n" {
		t.Fatalf("reordered v3 YAML output = %q", got)
	}
}
