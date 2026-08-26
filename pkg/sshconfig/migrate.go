package sshconfig

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

type legacyHost struct {
	Name   string            `json:"Name,omitempty" yaml:"Name,omitempty"`
	Notes  string            `json:"Notes,omitempty" yaml:"Notes,omitempty"`
	Data   map[string]string `json:"Data,omitempty" yaml:"-"`
	Config map[string]string `json:"-" yaml:"config,omitempty"`
}

type legacyGroup struct {
	Prefix string                `yaml:"Prefix,omitempty"`
	Common map[string]string     `yaml:"Common,omitempty"`
	Hosts  map[string]legacyHost `yaml:"Hosts,omitempty"`
}

type legacyYAML struct {
	Global  map[string]string      `yaml:"global,omitempty"`
	Default map[string]string      `yaml:"default,omitempty"`
	Groups  map[string]legacyGroup `yaml:",inline"`
}

// MigrateLegacyYAML converts the previous map-based YAML format into a v3
// document. It cannot recover ordering or repeated values already absent from
// the legacy input.
func MigrateLegacyYAML(data []byte, path string) (Schema, error) {
	var legacy legacyYAML
	if err := yaml.UnmarshalStrict(data, &legacy); err != nil {
		return Schema{}, fmt.Errorf("sshconfig: decode legacy YAML: %w", err)
	}
	var output bytes.Buffer
	if err := writeLegacyHost(&output, "*", "", legacy.Global); err != nil {
		return Schema{}, err
	}
	groupNames := sortedMapKeys(legacy.Groups)
	for _, groupName := range groupNames {
		group := legacy.Groups[groupName]
		hostNames := sortedMapKeys(group.Hosts)
		for _, hostName := range hostNames {
			host := group.Hosts[hostName]
			config := cloneStringMap(host.Config)
			if config == nil {
				config = make(map[string]string)
			}
			mergeMissing(config, group.Common)
			mergeMissing(config, legacy.Default)
			if err := writeLegacyHost(&output, group.Prefix+hostName, host.Notes, config); err != nil {
				return Schema{}, err
			}
		}
	}
	return schemaFromLegacyBytes(output.Bytes(), path)
}

// MigrateLegacyJSON converts the previous host-array JSON format into v3.
func MigrateLegacyJSON(data []byte, path string) (Schema, error) {
	var hosts []legacyHost
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&hosts); err != nil {
		return Schema{}, fmt.Errorf("sshconfig: decode legacy JSON: %w", err)
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return Schema{}, err
	}
	var output bytes.Buffer
	for _, host := range hosts {
		if err := writeLegacyHost(&output, host.Name, host.Notes, host.Data); err != nil {
			return Schema{}, err
		}
	}
	return schemaFromLegacyBytes(output.Bytes(), path)
}

func schemaFromLegacyBytes(data []byte, path string) (Schema, error) {
	doc, err := Parse(data)
	if err != nil {
		return Schema{}, err
	}
	schemaDocument, err := doc.ToSchema(path)
	if err != nil {
		return Schema{}, err
	}
	return NewSchema(schemaDocument), nil
}

func writeLegacyHost(output *bytes.Buffer, name, notes string, config map[string]string) error {
	if name == "" || len(config) == 0 {
		return nil
	}
	if err := ValidateDirectiveInput("Host", []string{name}, ""); err != nil {
		return fmt.Errorf("sshconfig: invalid legacy host %q: %w", name, err)
	}
	for _, note := range strings.Split(notes, "\n") {
		if err := ValidateDirectiveInput("Host", nil, note); err != nil {
			return fmt.Errorf("sshconfig: invalid legacy note: %w", err)
		}
	}
	for key, value := range config {
		if err := ValidateDirectiveInput(key, []string{value}, ""); err != nil {
			return fmt.Errorf("sshconfig: invalid legacy directive %q: %w", key, err)
		}
	}
	for _, note := range strings.Split(notes, "\n") {
		if note != "" {
			output.WriteString("# ")
			output.WriteString(note)
			output.WriteByte('\n')
		}
	}
	output.WriteString("Host ")
	output.WriteString(QuoteArgument(name))
	output.WriteByte('\n')
	for _, key := range sortedMapKeys(config) {
		output.WriteString("    ")
		output.WriteString(key)
		output.WriteByte(' ')
		output.WriteString(QuoteArgument(config[key]))
		output.WriteByte('\n')
	}
	return nil
}

func sortedMapKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func cloneStringMap(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}
	result := make(map[string]string, len(input))
	for key, value := range input {
		result[key] = value
	}
	return result
}

func mergeMissing(destination, source map[string]string) {
	for key, value := range source {
		if _, exists := destination[key]; !exists {
			destination[key] = value
		}
	}
}
