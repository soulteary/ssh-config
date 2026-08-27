package parser

import (
	"bytes"
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
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return sshconfig.Schema{}, false, nil
	}

	switch trimmed[0] {
	case '{':
		schema, err := sshconfig.UnmarshalSchemaJSON(data)
		return schema, true, err
	case '[':
		schema, err := sshconfig.MigrateLegacyJSON(data, "")
		return schema, true, err
	}

	if looksLikeStructuredYAML(trimmed) || strings.EqualFold(fileType, "YAML") && hasLegacyYAMLStructure(trimmed) {
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

func hasLegacyYAMLStructure(data []byte) bool {
	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil || len(document.Content) == 0 {
		return false
	}
	root := document.Content[0]
	if root.Kind != yaml.MappingNode {
		return false
	}
	for index := 0; index+1 < len(root.Content); index += 2 {
		group := root.Content[index+1]
		if group.Kind == yaml.AliasNode {
			group = group.Alias
		}
		if group == nil || group.Kind != yaml.MappingNode {
			continue
		}
		for field := 0; field+1 < len(group.Content); field += 2 {
			if group.Content[field].Value == "Hosts" {
				return true
			}
		}
	}
	return false
}

func looksLikeStructuredYAML(data []byte) bool {
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line == "---" || strings.HasPrefix(line, "#") {
			continue
		}
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "schemaversion:") ||
			strings.HasPrefix(lower, "documents:") ||
			strings.HasPrefix(lower, "global:") ||
			strings.HasPrefix(lower, "default:") ||
			(strings.HasPrefix(lower, "group ") && strings.Contains(line, ":")) {
			return true
		}
	}
	return false
}
