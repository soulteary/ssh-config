package parser

import (
	"fmt"
	"slices"

	Define "github.com/soulteary/ssh-config/internal/define"
	Fn "github.com/soulteary/ssh-config/internal/fn"
)

func ConvertToYAML(hostConfigs []Define.HostConfig) []byte {
	var output Define.YAMLOutput

	globalConfigs := Fn.FindGlobalConfig(hostConfigs)
	if len(globalConfigs) > 0 {
		output.Global = make(map[string]string)
		for _, config := range globalConfigs {
			for key, value := range config.Config {
				output.Global[key] = value
			}
		}
	}

	normalConfigs := Fn.FindNormalConfig(hostConfigs)
	if len(normalConfigs) > 0 {
		groups := make(map[string]Define.GroupConfig)

		for _, config := range normalConfigs {
			groupHostConfig := Define.HostConfig{}
			if config.Notes != "" {
				groupHostConfig.Notes = config.Notes
			}
			groupHostConfig.Config = config.Config

			groups[fmt.Sprintf("Group %s", config.Name)] = Define.GroupConfig{
				Common: make(map[string]string),
				Hosts: map[string]Define.HostConfig{
					config.Name: groupHostConfig,
				},
			}
		}

		output.Groups = groups
	}

	return Fn.GetYamlBytes(output)
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

			for hostName, originConfig := range groupConfig.Hosts {
				hostConfig := originConfig
				hostConfig.Name = hostName
				hostConfig.Extra.Prefix = prefix

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
	return hostConfigs
}
