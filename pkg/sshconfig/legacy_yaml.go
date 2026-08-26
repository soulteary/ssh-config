package sshconfig

import (
	"fmt"

	yaml "gopkg.in/yaml.v3"
)

type legacyYAMLNodeEntry struct {
	key   *yaml.Node
	value *yaml.Node
}

// ValidateLegacyYAML checks the resolved legacy schema while retaining enough
// source structure to reject duplicates inside merge operands.
func ValidateLegacyYAML(data []byte) error {
	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil {
		return err
	}
	if len(document.Content) == 0 {
		return nil
	}
	root := document.Content[0]
	if err := validateLegacyYAMLDuplicates(root, "document", make(map[*yaml.Node]bool)); err != nil {
		return err
	}
	return validateLegacyYAMLRoot(root)
}

func validateLegacyYAMLRoot(node *yaml.Node) error {
	entries, err := legacyYAMLMappingEntries(node, "document")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		switch entry.key.Value {
		case "global", "default":
			if err := validateLegacyYAMLStringMap(entry.value, entry.key.Value); err != nil {
				return err
			}
		default:
			if err := validateLegacyYAMLGroup(entry.value, entry.key.Value); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateLegacyYAMLGroup(node *yaml.Node, path string) error {
	entries, err := legacyYAMLMappingEntries(node, path)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		switch entry.key.Value {
		case "Prefix":
		case "Common":
			if err := validateLegacyYAMLStringMap(entry.value, path+".Common"); err != nil {
				return err
			}
		case "Hosts":
			if err := validateLegacyYAMLHosts(entry.value, path+".Hosts"); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown YAML field %q in %s", entry.key.Value, path)
		}
	}
	return nil
}

func validateLegacyYAMLHosts(node *yaml.Node, path string) error {
	entries, err := legacyYAMLMappingEntries(node, path)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := validateLegacyYAMLHost(entry.value, path+"."+entry.key.Value); err != nil {
			return err
		}
	}
	return nil
}

func validateLegacyYAMLHost(node *yaml.Node, path string) error {
	entries, err := legacyYAMLMappingEntries(node, path)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		switch entry.key.Value {
		case "Name", "Notes":
		case "config":
			if err := validateLegacyYAMLStringMap(entry.value, path+".config"); err != nil {
				return err
			}
		case "Extra":
			if err := validateLegacyYAMLExtra(entry.value, path+".Extra"); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown YAML field %q in %s", entry.key.Value, path)
		}
	}
	return nil
}

func validateLegacyYAMLExtra(node *yaml.Node, path string) error {
	entries, err := legacyYAMLMappingEntries(node, path)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.key.Value != "prefix" {
			return fmt.Errorf("unknown YAML field %q in %s", entry.key.Value, path)
		}
	}
	return nil
}

func validateLegacyYAMLStringMap(node *yaml.Node, path string) error {
	_, err := legacyYAMLMappingEntries(node, path)
	return err
}

func legacyYAMLMappingEntries(node *yaml.Node, path string) ([]legacyYAMLNodeEntry, error) {
	node = dereferenceLegacyYAMLAlias(node)
	if node == nil || node.Kind != yaml.MappingNode {
		return nil, nil // Typed decoding reports the more useful type error.
	}
	entries := resolveLegacyYAMLMapping(node, make(map[*yaml.Node]bool))
	for _, entry := range entries {
		if entry.key.Kind != yaml.ScalarNode || entry.key.Tag != "!!str" {
			return nil, fmt.Errorf("non-string YAML field in %s", path)
		}
	}
	return entries, nil
}

func validateLegacyYAMLDuplicates(node *yaml.Node, path string, visiting map[*yaml.Node]bool) error {
	node = dereferenceLegacyYAMLAlias(node)
	if node == nil || visiting[node] {
		return nil
	}
	visiting[node] = true
	defer delete(visiting, node)

	switch node.Kind {
	case yaml.MappingNode:
		seen := make(map[string]struct{}, len(node.Content)/2)
		for index := 0; index+1 < len(node.Content); index += 2 {
			key, value := node.Content[index], node.Content[index+1]
			name := key.Value
			if _, exists := seen[name]; exists {
				return fmt.Errorf("duplicate YAML field %q in %s", name, path)
			}
			seen[name] = struct{}{}
			childPath := path + "." + name
			if isLegacyYAMLMergeKey(key) {
				childPath = path + ".<<"
			}
			if err := validateLegacyYAMLDuplicates(value, childPath, visiting); err != nil {
				return err
			}
		}
	case yaml.SequenceNode:
		for index, child := range node.Content {
			if err := validateLegacyYAMLDuplicates(child, fmt.Sprintf("%s[%d]", path, index), visiting); err != nil {
				return err
			}
		}
	}
	return nil
}

func resolveLegacyYAMLMapping(node *yaml.Node, visiting map[*yaml.Node]bool) []legacyYAMLNodeEntry {
	node = dereferenceLegacyYAMLAlias(node)
	if node == nil || node.Kind != yaml.MappingNode || visiting[node] {
		return nil
	}
	visiting[node] = true
	defer delete(visiting, node)

	explicit := make(map[string]struct{})
	for index := 0; index+1 < len(node.Content); index += 2 {
		key := node.Content[index]
		if !isLegacyYAMLMergeKey(key) {
			explicit[key.Value] = struct{}{}
		}
	}

	result := make([]legacyYAMLNodeEntry, 0, len(node.Content)/2)
	seen := make(map[string]struct{})
	appendEntry := func(entry legacyYAMLNodeEntry, merged bool) {
		name := entry.key.Value
		if merged {
			if _, overridden := explicit[name]; overridden {
				return
			}
		}
		if _, exists := seen[name]; exists {
			return
		}
		seen[name] = struct{}{}
		result = append(result, entry)
	}

	for index := 0; index+1 < len(node.Content); index += 2 {
		key, value := node.Content[index], node.Content[index+1]
		if !isLegacyYAMLMergeKey(key) {
			appendEntry(legacyYAMLNodeEntry{key: key, value: value}, false)
			continue
		}
		for _, entry := range resolveLegacyYAMLMerge(value, visiting) {
			appendEntry(entry, true)
		}
	}
	return result
}

func resolveLegacyYAMLMerge(node *yaml.Node, visiting map[*yaml.Node]bool) []legacyYAMLNodeEntry {
	node = dereferenceLegacyYAMLAlias(node)
	if node == nil {
		return nil
	}
	if node.Kind == yaml.SequenceNode {
		var result []legacyYAMLNodeEntry
		seen := make(map[string]struct{})
		for _, item := range node.Content {
			for _, entry := range resolveLegacyYAMLMapping(item, visiting) {
				if _, exists := seen[entry.key.Value]; exists {
					continue
				}
				seen[entry.key.Value] = struct{}{}
				result = append(result, entry)
			}
		}
		return result
	}
	return resolveLegacyYAMLMapping(node, visiting)
}

func dereferenceLegacyYAMLAlias(node *yaml.Node) *yaml.Node {
	for node != nil && node.Kind == yaml.AliasNode {
		node = node.Alias
	}
	return node
}

func isLegacyYAMLMergeKey(node *yaml.Node) bool {
	return node != nil && (node.Tag == "!!merge" || node.Value == "<<")
}
