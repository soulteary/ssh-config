/**
 * Copyright 2024-2025 Su Yang (soulteary)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package parser

import (
	"fmt"
	"slices"

	Define "github.com/soulteary/ssh-config/v3/internal/define"
	Fn "github.com/soulteary/ssh-config/v3/internal/fn"
	"gopkg.in/yaml.v2"
	yamlv3 "gopkg.in/yaml.v3"
)

// mapToMapSlice 将 map[string]string 转为按 key 排序的 yaml.MapSlice，保证输出顺序稳定。
func mapToMapSlice(m map[string]string) yaml.MapSlice {
	if len(m) == 0 {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	out := make(yaml.MapSlice, 0, len(keys))
	for _, k := range keys {
		out = append(out, yaml.MapItem{Key: k, Value: m[k]})
	}
	return out
}

// hostConfigToMapSlice 将 HostConfig 转为固定字段顺序的 MapSlice（config 按 key 排序）。
// 键名使用小写 "config" 以与 yaml.v2 对无 tag 字段的默认 unmarshal 行为一致。
func hostConfigToMapSlice(c Define.HostConfig) yaml.MapSlice {
	items := yaml.MapSlice{}
	if c.Name != "" {
		items = append(items, yaml.MapItem{Key: "Name", Value: c.Name})
	}
	if c.Notes != "" {
		items = append(items, yaml.MapItem{Key: "Notes", Value: c.Notes})
	}
	// 仅当 Config 非 nil 时输出 config，以保证 round-trip 后 nil 仍为 nil、空 map 仍为空 map
	if c.Config != nil {
		items = append(items, yaml.MapItem{Key: "config", Value: mapToMapSlice(c.Config)})
	}
	if c.Extra.Prefix != "" {
		items = append(items, yaml.MapItem{Key: "Extra", Value: yaml.MapSlice{
			yaml.MapItem{Key: "Prefix", Value: c.Extra.Prefix},
		}})
	}
	return items
}

func ConvertToYAML(hostConfigs []Define.HostConfig) []byte {
	root := make(yaml.MapSlice, 0)

	globalConfigs := Fn.FindGlobalConfig(hostConfigs)
	if len(globalConfigs) > 0 {
		global := make(map[string]string)
		for _, config := range globalConfigs {
			for key, value := range config.Config {
				global[key] = value
			}
		}
		root = append(root, yaml.MapItem{Key: "global", Value: mapToMapSlice(global)})
	}

	normalConfigs := Fn.FindNormalConfig(hostConfigs)
	if len(normalConfigs) > 0 {
		usedGroupNames := make(map[string]bool)
		for _, config := range normalConfigs {
			groupName := uniqueLegacyGroupName(config, usedGroupNames)
			groupHostConfig := Define.HostConfig{}
			if config.Notes != "" {
				groupHostConfig.Notes = config.Notes
			}
			groupHostConfig.Config = config.Config
			hostConfig := hostConfigToMapSlice(groupHostConfig)
			groupItems := yaml.MapSlice{
				{Key: "Hosts", Value: yaml.MapSlice{
					{Key: config.Name, Value: hostConfig},
				}},
			}
			if config.Extra.Prefix != "" {
				groupItems = append(yaml.MapSlice{{Key: "Prefix", Value: config.Extra.Prefix}}, groupItems...)
			}
			root = append(root, yaml.MapItem{Key: groupName, Value: groupItems})
		}
	}

	yamlData, err := yaml.Marshal(root)
	if err != nil {
		fmt.Println("Error marshaling to YAML:", err)
		return nil
	}
	return yamlData
}

func uniqueLegacyGroupName(config Define.HostConfig, used map[string]bool) string {
	base := fmt.Sprintf("Group %s", config.Name)
	if !used[base] {
		used[base] = true
		return base
	}

	effective := fmt.Sprintf("Group %s%s", config.Extra.Prefix, config.Name)
	if !used[effective] {
		used[effective] = true
		return effective
	}
	for suffix := 2; ; suffix++ {
		candidate := fmt.Sprintf("%s (%d)", effective, suffix)
		if !used[candidate] {
			used[candidate] = true
			return candidate
		}
	}
}

type YAMLHostConfigGroup struct {
	Comments []string
	Config   map[string]string
}

func GroupYAMLConfig(input string) []Define.HostConfig {
	hostConfigs, _ := GroupYAMLConfigStrict(input)
	return hostConfigs
}

func GroupYAMLConfigStrict(input string) ([]Define.HostConfig, error) {
	yamlConfig, err := Fn.GetYamlDataStrict(input)
	if err != nil {
		return nil, fmt.Errorf("decode legacy YAML: %w", err)
	}

	var hostConfigs []Define.HostConfig

	if yamlConfig.Global != nil {
		hostConfig := Define.HostConfig{
			Name:   "*",
			Config: make(map[string]string),
		}
		for key, value := range yamlConfig.Global {
			hostConfig.Config[key] = value
		}
		hostConfigs = append(hostConfigs, hostConfig)
	}

	if yamlConfig.Groups != nil {
		groupOrder := legacyYAMLGroupOrder(input)
		keys := orderedLegacyKeys(groupOrder.groups, yamlConfig.Groups)

		for _, groupName := range keys {
			groupConfig := yamlConfig.Groups[groupName]

			prefix := ""
			if groupConfig.Prefix != "" {
				prefix = groupConfig.Prefix
			}

			hostNames := orderedLegacyKeys(groupOrder.hosts[groupName], groupConfig.Hosts)

			for _, hostName := range hostNames {
				originConfig := groupConfig.Hosts[hostName]
				hostConfig := originConfig
				hostConfig.Name = hostName
				hostConfig.Extra.Prefix = prefix
				if hostConfig.Config == nil && (len(groupConfig.Common) > 0 || len(yamlConfig.Default) > 0) {
					hostConfig.Config = make(map[string]string)
				}
				if hostConfig.Config != nil {
					if groupConfig.Common != nil {
						for key, value := range groupConfig.Common {

							if _, ok := hostConfig.Config[key]; !ok {
								hostConfig.Config[key] = value
							}
						}
					}

					if yamlConfig.Default != nil {
						for key, value := range yamlConfig.Default {
							if _, ok := hostConfig.Config[key]; !ok {
								hostConfig.Config[key] = value
							}
						}
					}
				}
				hostConfigs = append(hostConfigs, hostConfig)
			}
		}
	}
	if err := validateLegacyHostConfigs(hostConfigs); err != nil {
		return nil, err
	}
	return hostConfigs, nil
}

type yamlOrder struct {
	groups []string
	hosts  map[string][]string
}

func legacyYAMLGroupOrder(input string) yamlOrder {
	order := yamlOrder{hosts: make(map[string][]string)}
	var document yamlv3.Node
	if err := yamlv3.Unmarshal([]byte(input), &document); err != nil || len(document.Content) == 0 {
		return order
	}
	for _, group := range resolvedYAMLMapping(document.Content[0], make(map[*yamlv3.Node]bool)) {
		groupName := group.key.Value
		if groupName == "global" || groupName == "default" {
			continue
		}
		order.groups = append(order.groups, groupName)
		for _, field := range resolvedYAMLMapping(group.value, make(map[*yamlv3.Node]bool)) {
			if field.key.Value != "Hosts" {
				continue
			}
			for _, host := range resolvedYAMLMapping(field.value, make(map[*yamlv3.Node]bool)) {
				order.hosts[groupName] = append(order.hosts[groupName], host.key.Value)
			}
		}
	}
	return order
}

type yamlNodeEntry struct {
	key   *yamlv3.Node
	value *yamlv3.Node
}

func resolvedYAMLMapping(node *yamlv3.Node, visiting map[*yamlv3.Node]bool) []yamlNodeEntry {
	node = dereferenceYAMLAlias(node)
	if node == nil || node.Kind != yamlv3.MappingNode || visiting[node] {
		return nil
	}
	visiting[node] = true
	defer delete(visiting, node)

	explicit := make(map[string]struct{})
	for index := 0; index+1 < len(node.Content); index += 2 {
		key := node.Content[index]
		if !isYAMLMergeKey(key) {
			explicit[key.Value] = struct{}{}
		}
	}

	result := make([]yamlNodeEntry, 0, len(node.Content)/2)
	seen := make(map[string]struct{})
	appendEntry := func(entry yamlNodeEntry, merged bool) {
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
		if !isYAMLMergeKey(key) {
			appendEntry(yamlNodeEntry{key: key, value: value}, false)
			continue
		}
		for _, entry := range resolvedYAMLMerge(value, visiting) {
			appendEntry(entry, true)
		}
	}
	return result
}

func resolvedYAMLMerge(node *yamlv3.Node, visiting map[*yamlv3.Node]bool) []yamlNodeEntry {
	node = dereferenceYAMLAlias(node)
	if node == nil {
		return nil
	}
	if node.Kind == yamlv3.SequenceNode {
		var result []yamlNodeEntry
		seen := make(map[string]struct{})
		for _, item := range node.Content {
			for _, entry := range resolvedYAMLMapping(item, visiting) {
				if _, exists := seen[entry.key.Value]; exists {
					continue
				}
				seen[entry.key.Value] = struct{}{}
				result = append(result, entry)
			}
		}
		return result
	}
	return resolvedYAMLMapping(node, visiting)
}

func dereferenceYAMLAlias(node *yamlv3.Node) *yamlv3.Node {
	for node != nil && node.Kind == yamlv3.AliasNode {
		node = node.Alias
	}
	return node
}

func isYAMLMergeKey(node *yamlv3.Node) bool {
	return node != nil && node.Tag == "!!merge"
}

func orderedLegacyKeys[V any](preferred []string, values map[string]V) []string {
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
	slices.Sort(missing)
	return append(result, missing...)
}
