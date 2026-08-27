package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	Cmd "github.com/soulteary/ssh-config/v3/cmd"
	"github.com/soulteary/ssh-config/v3/pkg/sshconfig"
	"gopkg.in/yaml.v3"
)

// ProcessLossless converts SSH source or a v3 structured document without
// routing it through the legacy map representation.
func ProcessLossless(fileType, userInput string, args Cmd.Args) ([]byte, error) {
	data := []byte(userInput)
	schema, structured, err := decodeLosslessInput(fileType, data)
	if err != nil {
		return nil, err
	}
	if !structured {
		document, err := sshconfig.Parse(data)
		if err != nil {
			return nil, err
		}
		schemaDocument, err := document.ToSchema("")
		if err != nil {
			return nil, err
		}
		schema = sshconfig.NewSchema(schemaDocument)
	}

	switch {
	case args.ToYAML:
		return sshconfig.MarshalSchemaYAML(schema)
	case args.ToJSON:
		return sshconfig.MarshalSchemaJSON(schema)
	case args.ToSSH:
		document, err := schema.Document("")
		if err != nil {
			return nil, err
		}
		return document.MarshalPreserve()
	default:
		return nil, nil
	}
}

func decodeLosslessInput(fileType string, data []byte) (sshconfig.Schema, bool, error) {
	_ = fileType // Classification is based on the parsed document shape, not its filename.
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return sshconfig.Schema{}, false, nil
	}

	switch classifyStructuredInput(trimmed) {
	case structuredLegacyJSON:
		schema, err := sshconfig.MigrateLegacyJSON(data, "")
		return schema, true, err
	case structuredV3:
		if json.Valid(trimmed) {
			schema, err := sshconfig.UnmarshalSchemaJSON(data)
			return schema, true, err
		}
		schema, err := sshconfig.UnmarshalSchemaYAML(data)
		return schema, true, err
	case structuredLegacyYAML:
		schema, err := sshconfig.MigrateLegacyYAML(data, "")
		return schema, true, err
	case structuredYAMLUnknown:
		schema, schemaErr := sshconfig.UnmarshalSchemaYAML(data)
		if schemaErr == nil {
			return schema, true, nil
		}
		legacy, legacyErr := sshconfig.MigrateLegacyYAML(data, "")
		if legacyErr == nil {
			return legacy, true, nil
		}
		return sshconfig.Schema{}, true, fmt.Errorf("lossless input is neither v3 nor legacy YAML: v3: %v; legacy: %v", schemaErr, legacyErr)
	}

	return sshconfig.Schema{}, false, nil
}

type structuredInputKind uint8

const (
	structuredRaw structuredInputKind = iota
	structuredV3
	structuredLegacyJSON
	structuredLegacyYAML
	structuredYAMLUnknown
)

func classifyStructuredInput(data []byte) structuredInputKind {
	if len(data) == 0 {
		return structuredRaw
	}
	if data[0] == '[' {
		return structuredLegacyJSON
	}
	// A known OpenSSH directive at the start of a document is authoritative.
	// In particular, IgnoreUnknown may deliberately permit later lines whose
	// text happens to look like a v3 or legacy YAML field.
	if startsWithKnownSSHDirective(data) {
		return structuredRaw
	}

	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil || len(document.Content) == 0 {
		return structuredKindFromFirstLine(data)
	}
	root := document.Content[0]
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		root = root.Content[0]
	}
	for root != nil && root.Kind == yaml.AliasNode {
		root = root.Alias
	}
	if root == nil {
		return structuredRaw
	}
	if root.Kind != yaml.MappingNode {
		return structuredRaw
	}

	var hasV3, hasLegacy bool
	for index := 0; index+1 < len(root.Content); index += 2 {
		key := root.Content[index]
		if key.Kind != yaml.ScalarNode {
			continue
		}
		switch strings.ToLower(key.Value) {
		case "schemaversion", "documents":
			hasV3 = true
		case "global", "default":
			hasLegacy = true
		default:
			if strings.HasPrefix(strings.ToLower(key.Value), "group ") || mappingHasKey(root.Content[index+1], "Hosts") {
				hasLegacy = true
			}
		}
	}
	if hasV3 {
		return structuredV3
	}
	if hasLegacy {
		return structuredLegacyYAML
	}
	return structuredRaw
}

func mappingHasKey(node *yaml.Node, name string) bool {
	for node != nil && node.Kind == yaml.AliasNode {
		node = node.Alias
	}
	if node == nil || node.Kind != yaml.MappingNode {
		return false
	}
	for index := 0; index+1 < len(node.Content); index += 2 {
		if node.Content[index].Kind == yaml.ScalarNode && strings.EqualFold(node.Content[index].Value, name) {
			return true
		}
	}
	return false
}

func startsWithKnownSSHDirective(data []byte) bool {
	line := firstContentLine(data)
	if len(line) == 0 {
		return false
	}
	document, err := sshconfig.Parse(append(bytes.Clone(line), '\n'))
	if err != nil {
		return false
	}
	nodes := document.Nodes()
	if len(nodes) != 1 || nodes[0].Directive == nil {
		return false
	}
	_, known := sshconfig.LookupKeyword(nodes[0].Directive.KeywordValue)
	return known
}

func structuredKindFromFirstLine(data []byte) structuredInputKind {
	line := strings.TrimSpace(string(firstContentLine(data)))
	if line == "" {
		return structuredRaw
	}
	if strings.HasPrefix(line, "{") {
		return structuredYAMLUnknown
	}
	colon := strings.IndexByte(line, ':')
	if colon < 0 {
		return structuredRaw
	}
	key := strings.TrimSpace(line[:colon])
	key = strings.Trim(key, "\"'")
	switch strings.ToLower(key) {
	case "schemaversion", "documents":
		return structuredV3
	case "global", "default":
		return structuredLegacyYAML
	default:
		if strings.HasPrefix(strings.ToLower(key), "group ") {
			return structuredLegacyYAML
		}
	}
	return structuredRaw
}

func firstContentLine(data []byte) []byte {
	for _, line := range bytes.Split(data, []byte{'\n'}) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 || bytes.HasPrefix(line, []byte("#")) || bytes.Equal(line, []byte("---")) || bytes.HasPrefix(line, []byte("%YAML")) {
			continue
		}
		return line
	}
	return nil
}
