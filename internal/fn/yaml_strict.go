package fn

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

type yamlEntry struct {
	key   string
	value any
}

func validateLegacyYAML(data []byte) error {
	var ordered yaml.MapSlice
	if err := yaml.Unmarshal(data, &ordered); err != nil {
		return err
	}
	if err := validateYAMLRoot(ordered); err != nil {
		return err
	}

	// MapSlice retains explicit duplicates, but yaml.v2 omits values inherited
	// through merge keys from it. Validate the resolved map as a second pass so
	// merged unknown fields are checked without rejecting valid overrides.
	var resolved map[interface{}]interface{}
	if err := yaml.Unmarshal(data, &resolved); err != nil {
		return err
	}
	return validateYAMLRoot(resolved)
}

func validateYAMLRoot(value any) error {
	entries, err := yamlMappingEntries(value, "document")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		switch entry.key {
		case "global", "default":
			if err := validateYAMLStringMap(entry.value, entry.key); err != nil {
				return err
			}
		default:
			if err := validateYAMLGroup(entry.value, entry.key); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateYAMLGroup(value any, path string) error {
	entries, err := yamlMappingEntries(value, path)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		switch entry.key {
		case "Prefix":
		case "Common":
			if err := validateYAMLStringMap(entry.value, path+".Common"); err != nil {
				return err
			}
		case "Hosts":
			if err := validateYAMLHosts(entry.value, path+".Hosts"); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown YAML field %q in %s", entry.key, path)
		}
	}
	return nil
}

func validateYAMLHosts(value any, path string) error {
	entries, err := yamlMappingEntries(value, path)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := validateYAMLHost(entry.value, path+"."+entry.key); err != nil {
			return err
		}
	}
	return nil
}

func validateYAMLHost(value any, path string) error {
	entries, err := yamlMappingEntries(value, path)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		switch entry.key {
		case "Name", "Notes":
		case "config":
			if err := validateYAMLStringMap(entry.value, path+".config"); err != nil {
				return err
			}
		case "Extra":
			if err := validateYAMLExtra(entry.value, path+".Extra"); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown YAML field %q in %s", entry.key, path)
		}
	}
	return nil
}

func validateYAMLExtra(value any, path string) error {
	entries, err := yamlMappingEntries(value, path)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.key != "prefix" {
			return fmt.Errorf("unknown YAML field %q in %s", entry.key, path)
		}
	}
	return nil
}

func validateYAMLStringMap(value any, path string) error {
	_, err := yamlMappingEntries(value, path)
	return err
}

func yamlMappingEntries(value any, path string) ([]yamlEntry, error) {
	var entries []yamlEntry
	switch mapping := value.(type) {
	case nil:
		return nil, nil
	case yaml.MapSlice:
		entries = make([]yamlEntry, 0, len(mapping))
		for _, item := range mapping {
			key, ok := item.Key.(string)
			if !ok {
				return nil, fmt.Errorf("non-string YAML field in %s", path)
			}
			entries = append(entries, yamlEntry{key: key, value: item.Value})
		}
	case map[interface{}]interface{}:
		entries = make([]yamlEntry, 0, len(mapping))
		for rawKey, item := range mapping {
			key, ok := rawKey.(string)
			if !ok {
				return nil, fmt.Errorf("non-string YAML field in %s", path)
			}
			entries = append(entries, yamlEntry{key: key, value: item})
		}
	default:
		return nil, nil // The typed unmarshal reports the more useful type error.
	}

	seen := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		if _, exists := seen[entry.key]; exists {
			return nil, fmt.Errorf("duplicate YAML field %q in %s", entry.key, path)
		}
		seen[entry.key] = struct{}{}
	}
	return entries, nil
}
