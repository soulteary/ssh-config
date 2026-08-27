package sshconfig

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"
)

func TestSchemaJSONAndYAMLRoundTrip(t *testing.T) {
	t.Parallel()
	input := []byte("# note\r\nHost=example\r\n\tIdentityFile first\r\n\tIdentityFile second # backup\r\n")
	doc, _ := Parse(input)
	schemaDocument, err := doc.ToSchema("config")
	if err != nil {
		t.Fatal(err)
	}
	schema := NewSchema(schemaDocument)

	jsonData, err := MarshalSchemaJSON(schema)
	if err != nil {
		t.Fatal(err)
	}
	fromJSON, err := UnmarshalSchemaJSON(jsonData)
	if err != nil {
		t.Fatal(err)
	}
	assertSchemaDocumentBytes(t, fromJSON, "config", input)

	yamlData, err := MarshalSchemaYAML(schema)
	if err != nil {
		t.Fatal(err)
	}
	fromYAML, err := UnmarshalSchemaYAML(yamlData)
	if err != nil {
		t.Fatal(err)
	}
	assertSchemaDocumentBytes(t, fromYAML, "config", input)
}

func TestSchemaPreservesNonUTF8RawBytes(t *testing.T) {
	t.Parallel()
	input := []byte{'#', ' ', 0xff, '\n'}
	doc, _ := Parse(input)
	schemaDocument, err := doc.ToSchema("config")
	if err != nil {
		t.Fatal(err)
	}
	schema := NewSchema(schemaDocument)
	data, err := MarshalSchemaJSON(schema)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := UnmarshalSchemaJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	assertSchemaDocumentBytes(t, decoded, "config", input)
}

func TestSchemaRendersOnlyChangedDirective(t *testing.T) {
	t.Parallel()
	input := []byte("# untouched\r\n\tUser = old # account\r\nHost example\r\n")
	doc, _ := Parse(input)
	schemaDocument, _ := doc.ToSchema("config")
	schemaDocument.Nodes[1].Directive.Arguments = []string{"new user"}
	reconstructed, err := schemaDocument.Document()
	if err != nil {
		t.Fatal(err)
	}
	got, _ := reconstructed.MarshalPreserve()
	want := []byte("# untouched\r\nUser \"new user\" # account\r\nHost example\r\n")
	if !bytes.Equal(got, want) {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestSchemaPreservesInvalidNodesAsRawBytes(t *testing.T) {
	t.Parallel()
	input := []byte("Host example\n  ProxyCommand \"unterminated\n")
	doc, _ := Parse(input)
	if len(doc.Diagnostics()) == 0 {
		t.Fatal("test input did not produce a diagnostic")
	}

	schemaDocument, err := doc.ToSchema("config")
	if err != nil {
		t.Fatal(err)
	}
	if schemaDocument.Nodes[1].Type != "invalid" || schemaDocument.Nodes[1].Directive != nil {
		t.Fatalf("invalid schema node = %#v, want raw-only invalid node", schemaDocument.Nodes[1])
	}
	schema := NewSchema(schemaDocument)
	if err := schema.Validate(); err != nil {
		t.Fatalf("exported schema does not validate: %v", err)
	}

	jsonData, err := MarshalSchemaJSON(schema)
	if err != nil {
		t.Fatal(err)
	}
	fromJSON, err := UnmarshalSchemaJSON(jsonData)
	if err != nil {
		t.Fatal(err)
	}
	assertSchemaDocumentBytes(t, fromJSON, "config", input)

	yamlData, err := MarshalSchemaYAML(schema)
	if err != nil {
		t.Fatal(err)
	}
	fromYAML, err := UnmarshalSchemaYAML(yamlData)
	if err != nil {
		t.Fatal(err)
	}
	assertSchemaDocumentBytes(t, fromYAML, "config", input)
}

func TestSchemaPreservesUnsafeDirectiveViewsAsRawOnly(t *testing.T) {
	t.Parallel()
	tests := map[string][]byte{
		"hash in keyword":              []byte("Foo#Bar baz\n"),
		"vertical tab in keyword":      []byte("Foo\vBar baz\n"),
		"NUL in keyword":               []byte("Host\x00 example\n"),
		"NUL in unquoted argument":     []byte("Host example\x00hidden\n"),
		"NUL in trailing comment text": []byte("Host example # note\x00hidden\n"),
	}

	for name, input := range tests {
		name, input := name, input
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			document, err := Parse(input)
			if err != nil {
				t.Fatal(err)
			}
			schemaDocument, err := document.ToSchema("config")
			if err != nil {
				t.Fatal(err)
			}
			if len(schemaDocument.Nodes) != 1 || schemaDocument.Nodes[0].Type != "directive" || schemaDocument.Nodes[0].Directive != nil {
				t.Fatalf("schema node = %#v, want raw-only directive", schemaDocument.Nodes)
			}
			schema := NewSchema(schemaDocument)
			if err := schema.Validate(); err != nil {
				t.Fatalf("raw-only schema does not validate: %v", err)
			}

			jsonData, err := MarshalSchemaJSON(schema)
			if err != nil {
				t.Fatal(err)
			}
			fromJSON, err := UnmarshalSchemaJSON(jsonData)
			if err != nil {
				t.Fatal(err)
			}
			assertSchemaDocumentBytes(t, fromJSON, "config", input)

			yamlData, err := MarshalSchemaYAML(schema)
			if err != nil {
				t.Fatal(err)
			}
			fromYAML, err := UnmarshalSchemaYAML(yamlData)
			if err != nil {
				t.Fatal(err)
			}
			assertSchemaDocumentBytes(t, fromYAML, "config", input)
		})
	}
}

func TestSchemaEditPreservesMissingFinalNewline(t *testing.T) {
	t.Parallel()
	input := []byte("Host example\n  User old")
	doc, _ := Parse(input)
	schemaDocument, err := doc.ToSchema("config")
	if err != nil {
		t.Fatal(err)
	}
	schemaDocument.Nodes[1].Directive.Arguments = []string{"new"}

	reconstructed, err := schemaDocument.Document()
	if err != nil {
		t.Fatal(err)
	}
	got, err := reconstructed.MarshalPreserve()
	if err != nil {
		t.Fatal(err)
	}
	want := []byte("Host example\nUser new")
	if !bytes.Equal(got, want) {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestSchemaPreservesRawDirectiveWithoutArguments(t *testing.T) {
	t.Parallel()
	input := []byte("  FutureOption\n")
	doc, err := Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	schemaDocument, err := doc.ToSchema("config")
	if err != nil {
		t.Fatal(err)
	}
	reconstructed, err := schemaDocument.Document()
	if err != nil {
		t.Fatal(err)
	}
	got, err := reconstructed.MarshalPreserve()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, input) {
		t.Fatalf("schema round trip mismatch\n got: %q\nwant: %q", got, input)
	}
}

func TestSchemaNewDirectiveDefaultsToLF(t *testing.T) {
	t.Parallel()
	schemaDocument := SchemaDocument{Nodes: []SchemaNode{{
		Type:      "directive",
		Directive: &SchemaDirective{Keyword: "Host", Arguments: []string{"example"}},
	}}}
	doc, err := schemaDocument.Document()
	if err != nil {
		t.Fatal(err)
	}
	got, _ := doc.MarshalPreserve()
	if want := []byte("Host example\n"); !bytes.Equal(got, want) {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestSchemaDocumentRejectsInvalidNodeShapes(t *testing.T) {
	t.Parallel()
	tests := []SchemaDocument{
		{Nodes: []SchemaNode{{
			Type:      "comment",
			Directive: &SchemaDirective{Keyword: "Host", Arguments: []string{"example"}},
		}}},
		{Nodes: []SchemaNode{{
			Type:      "directive",
			Directive: &SchemaDirective{Keyword: "Host\nProxyCommand", Arguments: []string{"example"}},
		}}},
	}
	for _, schemaDocument := range tests {
		if _, err := schemaDocument.Document(); err == nil {
			t.Fatalf("SchemaDocument.Document() accepted invalid document: %#v", schemaDocument)
		}
	}
}

func TestSchemaValidationRejectsRawNodeShapeMismatches(t *testing.T) {
	t.Parallel()
	encode := func(raw string) string {
		return base64.StdEncoding.EncodeToString([]byte(raw))
	}
	tests := []struct {
		name string
		node SchemaNode
		want string
	}{
		{name: "declared comment", node: SchemaNode{Type: "comment", RawBase64: encode("Host example\n")}, want: "directive node, not comment"},
		{name: "declared blank", node: SchemaNode{Type: "blank", RawBase64: encode("User root\n")}, want: "directive node, not blank"},
		{name: "declared invalid", node: SchemaNode{Type: "invalid", RawBase64: encode("# comment\n")}, want: "comment node, not invalid"},
		{name: "multiple lines", node: SchemaNode{Type: "comment", RawBase64: encode("# cover\nProxyCommand command\n")}, want: "2 physical lines"},
		{name: "empty comment", node: SchemaNode{Type: "comment"}, want: "comment node has no raw bytes"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schema := NewSchema(SchemaDocument{Nodes: []SchemaNode{test.node}})
			err := schema.Validate()
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Validate() error = %v, want substring %q", err, test.want)
			}
		})
	}
}

func TestSchemaStrictValidation(t *testing.T) {
	t.Parallel()
	if _, err := UnmarshalSchemaJSON([]byte(`{"schemaVersion":3,"documents":[],"extra":true}`)); err == nil {
		t.Fatal("JSON accepted an unknown field")
	}
	if _, err := UnmarshalSchemaYAML([]byte("schemaVersion: 3\ndocuments: []\nextra: true\n")); err == nil {
		t.Fatal("YAML accepted an unknown field")
	}
	duplicateJSON := []string{
		`{"schemaVersion":2,"schemaVersion":3,"documents":[{"nodes":[]}]}`,
		`{"schemaVersion":3,"documents":[{"nodes":[{"type":"directive","directive":{"keyword":"ProxyCommand","keyword":"Host","arguments":["safe"]}}]}]}`,
	}
	for _, input := range duplicateJSON {
		if _, err := UnmarshalSchemaJSON([]byte(input)); err == nil {
			t.Fatalf("JSON accepted duplicate fields: %s", input)
		}
	}
	badVersion := Schema{SchemaVersion: 2, Documents: []SchemaDocument{{}}}
	if err := badVersion.Validate(); err == nil {
		t.Fatal("schema accepted an old version")
	}
	badBase64 := NewSchema(SchemaDocument{Nodes: []SchemaNode{{Type: "comment", RawBase64: "!"}}})
	if err := badBase64.Validate(); err == nil {
		t.Fatal("schema accepted invalid base64")
	}
	wrongNodeShape := NewSchema(SchemaDocument{Nodes: []SchemaNode{{
		Type:      "comment",
		Directive: &SchemaDirective{Keyword: "Host"},
	}}})
	if err := wrongNodeShape.Validate(); err == nil {
		t.Fatal("schema accepted a directive view on a comment node")
	}
}

func TestSchemaDocumentPathSelection(t *testing.T) {
	t.Parallel()

	single := NewSchema(SchemaDocument{})
	if _, err := single.Document(""); err != nil {
		t.Fatalf("single document with empty path was not selectable: %v", err)
	}

	for _, documents := range [][]SchemaDocument{
		{{}, {}},
		{{}, {Path: "config.d/work"}},
	} {
		schema := NewSchema(documents...)
		if err := schema.Validate(); err == nil || !strings.Contains(err.Error(), "empty path in a multi-document schema") {
			t.Fatalf("Validate() error = %v, want empty multi-document path error", err)
		}
	}

	multiple := NewSchema(
		SchemaDocument{Path: "config"},
		SchemaDocument{Path: "config.d/work"},
	)
	if _, err := multiple.Document(""); err == nil || !strings.Contains(err.Error(), "requires exactly one schema document") {
		t.Fatalf("Document(\"\") error = %v, want ambiguous selector error", err)
	}
}

func TestSchemaYAMLMergeExplicitValuesWin(t *testing.T) {
	t.Parallel()
	inputs := []string{
		"schemaVersion: 3\ndocuments:\n- nodes:\n  - type: directive\n    directive:\n      <<: {keyword: ProxyCommand, arguments: [evil]}\n      keyword: Host\n      arguments: [safe]\n",
		"schemaVersion: 3\ndocuments:\n- nodes:\n  - type: directive\n    directive:\n      keyword: Host\n      arguments: [safe]\n      <<: {keyword: ProxyCommand, arguments: [evil]}\n",
	}
	for _, input := range inputs {
		schema, err := UnmarshalSchemaYAML([]byte(input))
		if err != nil {
			t.Fatalf("UnmarshalSchemaYAML() error = %v", err)
		}
		doc, err := schema.Document("")
		if err != nil {
			t.Fatal(err)
		}
		output, err := doc.MarshalPreserve()
		if err != nil {
			t.Fatal(err)
		}
		if string(output) != "Host safe\n" {
			t.Fatalf("merged directive = %q, want Host safe", output)
		}
	}
}

func TestSchemaYAMLRejectsAmbiguousDocuments(t *testing.T) {
	t.Parallel()
	inputs := []string{
		"schemaVersion: 3\ndocuments:\n- nodes:\n  - type: directive\n    directive:\n      <<: {keyword: Host, keyword: Match, arguments: [safe]}\n",
		"schemaVersion: 3\ndocuments:\n- nodes: []\n---\nschemaVersion: 3\ndocuments:\n- nodes: []\n",
	}
	for _, input := range inputs {
		if _, err := UnmarshalSchemaYAML([]byte(input)); err == nil {
			t.Fatalf("UnmarshalSchemaYAML() accepted ambiguous input:\n%s", input)
		}
	}
}

func TestSchemaRejectsCrossLineDirectiveFields(t *testing.T) {
	t.Parallel()
	tests := []SchemaDirective{
		{Keyword: "Host\nProxyCommand", Arguments: []string{"example"}},
		{Keyword: "Host", Arguments: []string{"example\rUser root"}},
		{Keyword: "Host", Arguments: []string{"example"}, Comment: "# note\nProxyCommand command"},
	}
	for _, directive := range tests {
		directive := directive
		schema := NewSchema(SchemaDocument{Nodes: []SchemaNode{{Type: "directive", Directive: &directive}}})
		if err := schema.Validate(); err == nil {
			t.Fatalf("Validate() accepted cross-line directive: %+v", directive)
		}
	}
}

func TestMigrateLegacyFormats(t *testing.T) {
	t.Parallel()
	legacyYAML := []byte(`global:
  ServerAliveInterval: "30"
Group production:
  Prefix: corp-
  Common:
    User: deploy
  Hosts:
    api:
      Notes: main API
      config:
        HostName: api.example
`)
	schema, err := MigrateLegacyYAML(legacyYAML, "config")
	if err != nil {
		t.Fatalf("MigrateLegacyYAML() error = %v", err)
	}
	doc, err := schema.Document("config")
	if err != nil {
		t.Fatal(err)
	}
	yamlOutput, _ := doc.MarshalPreserve()
	for _, expected := range []string{"Host *", "ServerAliveInterval 30", "Host corp-api", "User deploy", "HostName api.example"} {
		if !strings.Contains(string(yamlOutput), expected) {
			t.Errorf("migrated YAML output missing %q:\n%s", expected, yamlOutput)
		}
	}

	legacyJSON := []byte(`[{"Name":"example","Notes":"note","Data":{"User":"root","IdentityFile":"~/.ssh/id"}}]`)
	schema, err = MigrateLegacyJSON(legacyJSON, "config")
	if err != nil {
		t.Fatalf("MigrateLegacyJSON() error = %v", err)
	}
	doc, _ = schema.Document("config")
	jsonOutput, _ := doc.MarshalPreserve()
	for _, expected := range []string{"# note", "Host example", "IdentityFile ~/.ssh/id", "User root"} {
		if !strings.Contains(string(jsonOutput), expected) {
			t.Errorf("migrated JSON output missing %q:\n%s", expected, jsonOutput)
		}
	}
}

func TestMigrateLegacyFormatsRejectCrossLineFields(t *testing.T) {
	t.Parallel()

	legacyJSON := []byte(`[{"Name":"example","Data":{"User":"alice\nHost injected"}}]`)
	if _, err := MigrateLegacyJSON(legacyJSON, "config"); err == nil {
		t.Fatal("MigrateLegacyJSON() accepted a cross-line value")
	}

	legacyYAML := []byte("Group test:\n  Hosts:\n    'example\nHost injected':\n      config:\n        User: alice\n")
	if _, err := MigrateLegacyYAML(legacyYAML, "config"); err == nil {
		t.Fatal("MigrateLegacyYAML() accepted a cross-line host")
	}
}

func TestMigrateLegacyYAMLAllowsMergeOverrides(t *testing.T) {
	t.Parallel()
	inputs := []string{
		"global:\n  <<: &defaults\n    User: alice\n    Port: \"22\"\n  User: bob\n",
		"global:\n  User: bob\n  <<: &defaults\n    User: alice\n    Port: \"22\"\n",
	}
	for _, input := range inputs {
		schema, err := MigrateLegacyYAML([]byte(input), "config")
		if err != nil {
			t.Fatalf("MigrateLegacyYAML() error = %v", err)
		}
		doc, err := schema.Document("config")
		if err != nil {
			t.Fatal(err)
		}
		output, err := doc.MarshalPreserve()
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{"Port 22", "User bob"} {
			if !strings.Contains(string(output), want) {
				t.Fatalf("migrated output missing %q:\n%s", want, output)
			}
		}
	}
}

func TestMigrateLegacyYAMLRejectsDuplicateMergeOperands(t *testing.T) {
	t.Parallel()
	input := []byte("global:\n  <<: {User: alice, User: bob}\n")
	if _, err := MigrateLegacyYAML(input, "config"); err == nil {
		t.Fatal("MigrateLegacyYAML() accepted duplicate fields inside a merge operand")
	}
}

func TestMigrateLegacyYAMLRejectsQuotedMergeKeyAsUnknownField(t *testing.T) {
	t.Parallel()
	input := []byte("Group work:\n  \"<<\": {}\n")
	if _, err := MigrateLegacyYAML(input, "config"); err == nil || !strings.Contains(err.Error(), "unknown YAML field") {
		t.Fatalf("MigrateLegacyYAML() error = %v, want unknown field error", err)
	}
}

func TestMigrateLegacyYAMLPreservesGroupAndHostOrder(t *testing.T) {
	t.Parallel()
	input := []byte(`Group z:
  Hosts:
    special.example.com:
      config:
        User: special
    '*.example.com':
      config:
        User: wildcard
Group a:
  Hosts:
    last.example.com:
      config:
        User: last
`)
	schema, err := MigrateLegacyYAML(input, "config")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := schema.Document("config")
	if err != nil {
		t.Fatal(err)
	}
	output, err := doc.MarshalPreserve()
	if err != nil {
		t.Fatal(err)
	}
	special := strings.Index(string(output), "Host special.example.com")
	wildcard := strings.Index(string(output), "Host *.example.com")
	last := strings.Index(string(output), "Host last.example.com")
	if special < 0 || wildcard < 0 || last < 0 || !(special < wildcard && wildcard < last) {
		t.Fatalf("migration changed group or host order:\n%s", output)
	}
}

func TestMigrateLegacyYAMLPreservesMergedHostOrder(t *testing.T) {
	t.Parallel()
	input := []byte(`Group template: &group
  Prefix: template-
  Hosts:
    z-host:
      config:
        User: z
    a-host:
      config:
        User: a
Group inherited:
  <<: *group
  Prefix: inherited-
`)
	schema, err := MigrateLegacyYAML(input, "config")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := schema.Document("config")
	if err != nil {
		t.Fatal(err)
	}
	output, err := doc.MarshalPreserve()
	if err != nil {
		t.Fatal(err)
	}
	z := strings.Index(string(output), "Host inherited-z-host")
	a := strings.Index(string(output), "Host inherited-a-host")
	if z < 0 || a < 0 || z > a {
		t.Fatalf("migration changed inherited host order:\n%s", output)
	}
}

func assertSchemaDocumentBytes(t *testing.T, schema Schema, path string, want []byte) {
	t.Helper()
	doc, err := schema.Document(path)
	if err != nil {
		t.Fatal(err)
	}
	got, err := doc.MarshalPreserve()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("schema round trip mismatch\n got: %q\nwant: %q", got, want)
	}
}
