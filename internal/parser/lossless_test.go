package parser_test

import (
	"bytes"
	"strings"
	"testing"

	Cmd "github.com/soulteary/ssh-config/v3/cmd"
	Parser "github.com/soulteary/ssh-config/v3/internal/parser"
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

func TestProcessLosslessPreservesCommentOnlySSHInput(t *testing.T) {
	t.Parallel()
	original := []byte("# keep me\r\n# second\r\n")

	// Legacy format detection classifies a comment-only document as YAML.
	// Default lossless conversion must still prefer the original SSH bytes
	// when the input has no v3 or legacy YAML structure markers.
	structured, err := Parser.ProcessLossless("YAML", string(original), Cmd.Args{ToYAML: true})
	if err != nil {
		t.Fatal(err)
	}
	reconstructed, err := Parser.ProcessLossless("YAML", string(structured), Cmd.Args{ToSSH: true})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(reconstructed, original) {
		t.Fatalf("comment-only round trip = %q, want %q", reconstructed, original)
	}
}

func TestProcessLosslessMigratesCustomLegacyYAMLGroup(t *testing.T) {
	t.Parallel()
	input := `work:
  Hosts:
    example:
      config:
        User: alice
`

	got, err := Parser.ProcessLossless("YAML", input, Cmd.Args{ToSSH: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Host example", "User alice"} {
		if !bytes.Contains(got, []byte(want)) {
			t.Fatalf("migrated SSH = %q, want it to contain %q", got, want)
		}
	}
}

func TestProcessLosslessPreservesSSHThatLooksLikeYAML(t *testing.T) {
	t.Parallel()
	input := []byte("RemoteCommand echo foo: {}\n")

	got, err := Parser.ProcessLossless("YAML", string(input), Cmd.Args{ToSSH: true})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, input) {
		t.Fatalf("YAML-like SSH = %q, want %q", got, input)
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

func TestProcessLosslessClassifiesMalformedStructuredPrefixes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "schema version", input: "schemaVersion: [\n", want: "decode v3 YAML"},
		{name: "documents", input: "documents: [\n", want: "decode v3 YAML"},
		{name: "quoted documents", input: "\"documents\": [\n", want: "decode v3 YAML"},
		{name: "global", input: "global: [\n", want: "decode legacy YAML"},
		{name: "default", input: "default: [\n", want: "decode legacy YAML"},
		{name: "group", input: "Group work: [\n", want: "decode legacy YAML"},
		{name: "flow mapping", input: "{broken\n", want: "neither v3 nor legacy YAML"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := Parser.ProcessLossless("YAML", test.input, Cmd.Args{ToSSH: true})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("ProcessLossless() error = %v, want substring %q", err, test.want)
			}
		})
	}
}

func TestProcessLosslessSelectsSchemaDocumentByPath(t *testing.T) {
	t.Parallel()
	input := `schemaVersion: 3
documents:
  - path: /home/alice/.ssh/config
    nodes:
      - type: directive
        directive:
          keyword: Host
          arguments: [alice]
  - path: /home/bob/.ssh/config
    nodes:
      - type: directive
        directive:
          keyword: Host
          arguments: [bob]
`

	got, err := Parser.ProcessLossless("YAML", input, Cmd.Args{
		ToSSH:        true,
		DocumentPath: "/home/bob/.ssh/config",
	})
	if err != nil {
		t.Fatal(err)
	}
	if want := "Host bob\n"; string(got) != want {
		t.Fatalf("selected document = %q, want %q", got, want)
	}
	if _, err := Parser.ProcessLossless("YAML", input, Cmd.Args{ToSSH: true}); err == nil {
		t.Fatal("multi-document schema was accepted without -document-path")
	}
	if _, err := Parser.ProcessLossless("YAML", input, Cmd.Args{ToSSH: true, DocumentPath: "missing"}); err == nil {
		t.Fatal("unknown document path was accepted")
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

func TestProcessLosslessPreservesIgnoreUnknownMarkerCollisions(t *testing.T) {
	t.Parallel()
	input := "IgnoreUnknown schemaVersion:,documents:,global:,default:\n" +
		"schemaVersion: 3\n" +
		"documents: keep\n" +
		"global: keep\n" +
		"default: keep\n" +
		"Host *\n    User audit\n"

	got, err := Parser.ProcessLossless("YAML", input, Cmd.Args{ToSSH: true})
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != input {
		t.Fatalf("marker-bearing SSH = %q, want %q", got, input)
	}
}

func TestProcessLosslessAcceptsFlowAndQuotedYAML(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "flow legacy",
			input: `{global: {User: alice}}`,
			want:  "Host *\n    User alice\n",
		},
		{
			name:  "JSON-shaped legacy YAML",
			input: `{"global":{"User":"alice"}}`,
			want:  "Host *\n    User alice\n",
		},
		{
			name:  "flow v3",
			input: `{schemaVersion: 3, documents: [{nodes: [{type: directive, directive: {keyword: Host, arguments: [example]}}]}]}`,
			want:  "Host example\n",
		},
		{
			name: "quoted v3 keys",
			input: `"schemaVersion": 3
"documents":
  - nodes:
      - type: directive
        directive:
          keyword: Host
          arguments: [example]
`,
			want: "Host example\n",
		},
		{
			name: "quoted legacy key",
			input: `"global":
  User: alice
`,
			want: "Host *\n    User alice\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := Parser.ProcessLossless("YAML", test.input, Cmd.Args{ToSSH: true})
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != test.want {
				t.Fatalf("output = %q, want %q", got, test.want)
			}
		})
	}
}
