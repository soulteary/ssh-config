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

	Define "github.com/soulteary/ssh-config/internal/define"
	Fn "github.com/soulteary/ssh-config/internal/fn"
	"gopkg.in/yaml.v2"
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
		groupNames := make([]string, 0, len(normalConfigs))
		groupsData := make(map[string]yaml.MapSlice)
		for _, config := range normalConfigs {
			groupName := fmt.Sprintf("Group %s", config.Name)
			groupNames = append(groupNames, groupName)
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
			groupsData[groupName] = groupItems
		}
		slices.Sort(groupNames)
		for _, groupName := range groupNames {
			root = append(root, yaml.MapItem{Key: groupName, Value: groupsData[groupName]})
		}
	}

	yamlData, err := yaml.Marshal(root)
	if err != nil {
		fmt.Println("Error marshaling to YAML:", err)
		return nil
	}
	return yamlData
}

type YAMLHostConfigGroup struct {
	Comments []string
	Config   map[string]string
}

func GroupYAMLConfig(input string) []Define.HostConfig {
	yamlConfig := Fn.GetYamlData(input)

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
		keys := make([]string, 0)
		for key := range yamlConfig.Groups {
			keys = append(keys, key)
		}
		slices.Sort(keys)

		for _, groupName := range keys {
			groupConfig := yamlConfig.Groups[groupName]

			prefix := ""
			if groupConfig.Prefix != "" {
				prefix = groupConfig.Prefix
			}

			hostNames := make([]string, 0, len(groupConfig.Hosts))
			for hostName := range groupConfig.Hosts {
				hostNames = append(hostNames, hostName)
			}
			slices.Sort(hostNames)

			for _, hostName := range hostNames {
				originConfig := groupConfig.Hosts[hostName]
				hostConfig := originConfig
				hostConfig.Name = hostName
				hostConfig.Extra.Prefix = prefix
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
					hostConfigs = append(hostConfigs, hostConfig)
				}
			}
		}
	}
	return hostConfigs
}
