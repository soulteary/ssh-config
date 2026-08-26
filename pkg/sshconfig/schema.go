package sshconfig

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"unicode/utf8"

	"gopkg.in/yaml.v2"
)

// SchemaVersion is the lossless structured configuration format version.
const SchemaVersion = 3

// Schema is the root of the v3 YAML/JSON representation.
type Schema struct {
	SchemaVersion int              `json:"schemaVersion" yaml:"schemaVersion"`
	Documents     []SchemaDocument `json:"documents" yaml:"documents"`
}

// SchemaDocument is one physical SSH configuration file.
type SchemaDocument struct {
	Path  string       `json:"path,omitempty" yaml:"path,omitempty"`
	Nodes []SchemaNode `json:"nodes" yaml:"nodes"`
}

// SchemaNode preserves one physical source line. RawBase64 is authoritative
// while its parsed Directive remains unchanged. If Directive is edited, the
// importer renders only that node canonically.
type SchemaNode struct {
	Type      string           `json:"type" yaml:"type"`
	RawBase64 string           `json:"rawBase64,omitempty" yaml:"rawBase64,omitempty"`
	Directive *SchemaDirective `json:"directive,omitempty" yaml:"directive,omitempty"`
}

// SchemaDirective is an editable semantic view of a directive node.
type SchemaDirective struct {
	Keyword   string   `json:"keyword" yaml:"keyword"`
	Arguments []string `json:"arguments,omitempty" yaml:"arguments,omitempty"`
	Comment   string   `json:"comment,omitempty" yaml:"comment,omitempty"`
}

// ToSchema exports a document without discarding its original source bytes.
func (d *Document) ToSchema(path string) (SchemaDocument, error) {
	serialized, err := d.MarshalPreserve()
	if err != nil {
		return SchemaDocument{}, err
	}
	clean, err := Parse(serialized)
	if err != nil {
		return SchemaDocument{}, err
	}
	result := SchemaDocument{Path: path, Nodes: make([]SchemaNode, 0, len(clean.nodes))}
	for _, node := range clean.nodes {
		raw := clean.source[node.Span.Start:node.Span.End]
		schemaNode := SchemaNode{
			Type:      schemaNodeType(node.Kind),
			RawBase64: base64.StdEncoding.EncodeToString(raw),
		}
		// Invalid nodes may carry a partially parsed Directive internally so
		// diagnostics can point at tokens. They must remain raw-only in the
		// schema: exposing an incomplete editable view violates the node shape
		// contract and can discard the malformed bytes when reconstructed.
		if node.Kind == NodeDirective && node.Directive != nil && utf8.Valid(raw) {
			directive := node.Directive
			arguments := make([]string, 0, len(directive.Arguments))
			for _, argument := range directive.Arguments {
				arguments = append(arguments, argument.Value)
			}
			view := &SchemaDirective{
				Keyword:   string(clean.source[directive.Keyword.Span.Start:directive.Keyword.Span.End]),
				Arguments: arguments,
			}
			if directive.Comment != nil {
				view.Comment = string(clean.source[directive.Comment.Span.Start:directive.Comment.Span.End])
			}
			schemaNode.Directive = view
		}
		result.Nodes = append(result.Nodes, schemaNode)
	}
	return result, nil
}

// NewSchema creates a v3 schema from one or more documents.
func NewSchema(documents ...SchemaDocument) Schema {
	return Schema{SchemaVersion: SchemaVersion, Documents: documents}
}

// Document reconstructs one schema document. An empty path selects the first
// document when the schema contains exactly one document.
func (s Schema) Document(path string) (*Document, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	for _, document := range s.Documents {
		if document.Path == path || path == "" && len(s.Documents) == 1 {
			return document.Document()
		}
	}
	return nil, fmt.Errorf("sshconfig: schema document %q not found", path)
}

// Document reconstructs a lossless syntax document from a v3 document.
func (s SchemaDocument) Document() (*Document, error) {
	var output bytes.Buffer
	for index, node := range s.Nodes {
		raw, err := decodeSchemaRaw(node.RawBase64)
		if err != nil {
			return nil, fmt.Errorf("sshconfig: schema node %d: %w", index, err)
		}
		if node.Directive == nil {
			if len(raw) == 0 && node.Type != "blank" {
				return nil, fmt.Errorf("sshconfig: schema node %d has neither raw bytes nor a directive", index)
			}
			output.Write(raw)
			continue
		}
		if schemaRawMatchesDirective(raw, node.Directive) {
			output.Write(raw)
			continue
		}
		output.Write(renderSchemaDirective(node.Directive, detectLineEnding(raw), len(raw) == 0))
	}
	return Parse(output.Bytes())
}

// Validate checks the v3 envelope and node shapes.
func (s Schema) Validate() error {
	if s.SchemaVersion != SchemaVersion {
		return fmt.Errorf("sshconfig: unsupported schema version %d", s.SchemaVersion)
	}
	if len(s.Documents) == 0 {
		return fmt.Errorf("sshconfig: schema has no documents")
	}
	seen := make(map[string]bool)
	for index, document := range s.Documents {
		if document.Path != "" {
			if seen[document.Path] {
				return fmt.Errorf("sshconfig: duplicate schema document path %q", document.Path)
			}
			seen[document.Path] = true
		}
		for nodeIndex, node := range document.Nodes {
			switch node.Type {
			case "blank", "comment", "directive", "invalid":
			default:
				return fmt.Errorf("sshconfig: document %d node %d has unknown type %q", index, nodeIndex, node.Type)
			}
			if node.Directive != nil && node.Directive.Keyword == "" {
				return fmt.Errorf("sshconfig: document %d node %d has an empty keyword", index, nodeIndex)
			}
			if node.Type == "directive" && node.Directive == nil && node.RawBase64 == "" {
				return fmt.Errorf("sshconfig: document %d node %d has no directive or raw bytes", index, nodeIndex)
			}
			if node.Type != "directive" && node.Directive != nil {
				return fmt.Errorf("sshconfig: document %d node %d of type %q has a directive", index, nodeIndex, node.Type)
			}
			if _, err := decodeSchemaRaw(node.RawBase64); err != nil {
				return fmt.Errorf("sshconfig: document %d node %d: %w", index, nodeIndex, err)
			}
		}
	}
	return nil
}

// MarshalSchemaJSON serializes a validated schema as indented JSON.
func MarshalSchemaJSON(schema Schema) ([]byte, error) {
	if err := schema.Validate(); err != nil {
		return nil, err
	}
	return json.MarshalIndent(schema, "", "  ")
}

// UnmarshalSchemaJSON strictly decodes one JSON schema document.
func UnmarshalSchemaJSON(data []byte) (Schema, error) {
	var schema Schema
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&schema); err != nil {
		return Schema{}, fmt.Errorf("sshconfig: decode v3 JSON: %w", err)
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return Schema{}, err
	}
	if err := schema.Validate(); err != nil {
		return Schema{}, err
	}
	return schema, nil
}

// MarshalSchemaYAML serializes a validated schema as YAML.
func MarshalSchemaYAML(schema Schema) ([]byte, error) {
	if err := schema.Validate(); err != nil {
		return nil, err
	}
	return yaml.Marshal(schema)
}

// UnmarshalSchemaYAML strictly decodes one YAML schema document.
func UnmarshalSchemaYAML(data []byte) (Schema, error) {
	var schema Schema
	if err := yaml.UnmarshalStrict(data, &schema); err != nil {
		return Schema{}, fmt.Errorf("sshconfig: decode v3 YAML: %w", err)
	}
	if err := schema.Validate(); err != nil {
		return Schema{}, err
	}
	return schema, nil
}

func ensureJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err == io.EOF {
		return nil
	} else if err != nil {
		return fmt.Errorf("sshconfig: decode trailing JSON: %w", err)
	}
	return fmt.Errorf("sshconfig: multiple JSON values are not allowed")
}

func schemaNodeType(kind NodeKind) string {
	switch kind {
	case NodeBlank:
		return "blank"
	case NodeComment:
		return "comment"
	case NodeDirective:
		return "directive"
	default:
		return "invalid"
	}
}

func decodeSchemaRaw(encoded string) ([]byte, error) {
	if encoded == "" {
		return nil, nil
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid rawBase64: %w", err)
	}
	return decoded, nil
}

func schemaRawMatchesDirective(raw []byte, expected *SchemaDirective) bool {
	if len(raw) == 0 {
		return false
	}
	doc, err := Parse(raw)
	if err != nil || len(doc.nodes) != 1 || doc.nodes[0].Directive == nil {
		return false
	}
	directive := doc.nodes[0].Directive
	actual := &SchemaDirective{
		Keyword: string(doc.source[directive.Keyword.Span.Start:directive.Keyword.Span.End]),
	}
	for _, argument := range directive.Arguments {
		actual.Arguments = append(actual.Arguments, argument.Value)
	}
	if directive.Comment != nil {
		actual.Comment = string(doc.source[directive.Comment.Span.Start:directive.Comment.Span.End])
	}
	return reflect.DeepEqual(actual, expected)
}

func renderSchemaDirective(directive *SchemaDirective, lineEnding string, defaultLineEnding bool) []byte {
	if lineEnding == "" && defaultLineEnding {
		lineEnding = "\n"
	}
	var output bytes.Buffer
	output.WriteString(directive.Keyword)
	if len(directive.Arguments) > 0 {
		output.WriteByte(' ')
		writeArguments(&output, directive.Arguments)
	}
	if directive.Comment != "" {
		output.WriteByte(' ')
		if directive.Comment[0] != '#' {
			output.WriteByte('#')
			output.WriteByte(' ')
		}
		output.WriteString(directive.Comment)
	}
	output.WriteString(lineEnding)
	return output.Bytes()
}

func detectLineEnding(raw []byte) string {
	if bytes.HasSuffix(raw, []byte("\r\n")) {
		return "\r\n"
	}
	if bytes.HasSuffix(raw, []byte("\n")) {
		return "\n"
	}
	if bytes.HasSuffix(raw, []byte("\r")) {
		return "\r"
	}
	return ""
}
