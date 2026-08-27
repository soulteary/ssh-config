package sshconfig

import (
	"bytes"
	jsonv2 "encoding/json/v2"
	"fmt"
	"sort"
	"strings"

	yaml "gopkg.in/yaml.v3"
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
	if err := ValidateLegacyYAML(data); err != nil {
		return Schema{}, fmt.Errorf("sshconfig: decode legacy YAML: %w", err)
	}
	order, err := legacyYAMLSourceOrder(data)
	if err != nil {
		return Schema{}, fmt.Errorf("sshconfig: decode legacy YAML order: %w", err)
	}
	if err := yaml.Unmarshal(data, &legacy); err != nil {
		return Schema{}, fmt.Errorf("sshconfig: decode legacy YAML: %w", err)
	}
	var output bytes.Buffer
	if len(legacy.Global) > 0 {
		if err := writeLegacyHost(&output, "*", "", legacy.Global); err != nil {
			return Schema{}, err
		}
	}
	groupNames := orderedLegacyMapKeys(order.groups, legacy.Groups)
	for _, groupName := range groupNames {
		group := legacy.Groups[groupName]
		hostNames := orderedLegacyMapKeys(order.hosts[groupName], group.Hosts)
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

type legacyYAMLOrder struct {
	groups []string
	hosts  map[string][]string
}

func legacyYAMLSourceOrder(data []byte) (legacyYAMLOrder, error) {
	order := legacyYAMLOrder{hosts: make(map[string][]string)}
	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil {
		return order, err
	}
	if len(document.Content) == 0 {
		return order, nil
	}
	groups, err := legacyYAMLMappingEntries(document.Content[0], "document")
	if err != nil {
		return order, err
	}
	for _, group := range groups {
		groupName := group.key.Value
		if groupName == "global" || groupName == "default" {
			continue
		}
		order.groups = append(order.groups, groupName)
		fields, err := legacyYAMLMappingEntries(group.value, groupName)
		if err != nil {
			return order, err
		}
		for _, field := range fields {
			if field.key.Value != "Hosts" {
				continue
			}
			hosts, err := legacyYAMLMappingEntries(field.value, groupName+".Hosts")
			if err != nil {
				return order, err
			}
			for _, host := range hosts {
				order.hosts[groupName] = append(order.hosts[groupName], host.key.Value)
			}
		}
	}
	return order, nil
}

func orderedLegacyMapKeys[V any](preferred []string, values map[string]V) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, key := range preferred {
		if _, exists := values[key]; !exists {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, key)
	}
	missing := make([]string, 0, len(values)-len(result))
	for key := range values {
		if _, exists := seen[key]; !exists {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	return append(result, missing...)
}

// MigrateLegacyJSON converts the previous host-array JSON format into v3.
func MigrateLegacyJSON(data []byte, path string) (Schema, error) {
	var hosts []legacyHost
	if err := jsonv2.Unmarshal(data, &hosts,
		jsonv2.RejectUnknownMembers(true),
		jsonv2.MatchCaseInsensitiveNames(true),
	); err != nil {
		return Schema{}, fmt.Errorf("sshconfig: decode legacy JSON: %w", err)
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
	if name == "" {
		return fmt.Errorf("sshconfig: legacy host name is empty")
	}
	if err := ValidateDirectiveInput("Host", []string{name}, ""); err != nil {
		return fmt.Errorf("sshconfig: invalid legacy host %q: %w", name, err)
	}
	hostArguments, err := parseLegacyArgumentText("Host", name)
	if err != nil {
		return fmt.Errorf("sshconfig: invalid legacy host %q: %w", name, err)
	}
	if len(hostArguments) == 0 {
		return fmt.Errorf("sshconfig: invalid legacy host %q: Host requires at least one pattern", name)
	}
	for _, note := range strings.Split(notes, "\n") {
		if err := ValidateDirectiveInput("Host", nil, note); err != nil {
			return fmt.Errorf("sshconfig: invalid legacy note: %w", err)
		}
	}
	argumentsByKey := make(map[string][]string, len(config))
	for key, value := range config {
		if err := ValidateDirectiveInput(key, []string{value}, ""); err != nil {
			return fmt.Errorf("sshconfig: invalid legacy directive %q: %w", key, err)
		}
		arguments, err := parseLegacyArgumentText(key, value)
		if err != nil {
			return fmt.Errorf("sshconfig: invalid legacy directive %q: %w", key, err)
		}
		argumentsByKey[key] = arguments
	}
	for _, note := range strings.Split(notes, "\n") {
		if note != "" {
			output.WriteString("# ")
			output.WriteString(note)
			output.WriteByte('\n')
		}
	}
	output.WriteString("Host ")
	writeArguments(output, hostArguments)
	output.WriteByte('\n')
	for _, key := range sortedMapKeys(config) {
		output.WriteString("    ")
		output.WriteString(key)
		if len(argumentsByKey[key]) > 0 {
			output.WriteByte(' ')
			writeArguments(output, argumentsByKey[key])
		}
		output.WriteByte('\n')
	}
	return nil
}

func parseLegacyArgumentText(keyword, value string) ([]string, error) {
	line := keyword
	if value != "" {
		line += " " + value
	}
	document, err := Parse([]byte(line + "\n"))
	if err != nil {
		return nil, err
	}
	if diagnostics := document.Diagnostics(); len(diagnostics) > 0 {
		return nil, fmt.Errorf("%s", diagnostics[0].Message)
	}
	nodes := document.Nodes()
	if len(nodes) != 1 || nodes[0].Kind != NodeDirective || nodes[0].Directive == nil {
		return nil, fmt.Errorf("cannot parse legacy argument text")
	}
	arguments := make([]string, 0, len(nodes[0].Directive.Arguments))
	for _, argument := range nodes[0].Directive.Arguments {
		arguments = append(arguments, argument.Value)
	}
	return arguments, nil
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
