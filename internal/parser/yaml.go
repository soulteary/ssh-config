package parser

import (
	"fmt"

	Define "github.com/soulteary/ssh-yaml/internal/define"
	Fn "github.com/soulteary/ssh-yaml/internal/fn"
)

type GroupConfig struct {
	Prefix string                       `yaml:"Prefix,omitempty"`
	Config Define.HostConfig            `yaml:"Config,omitempty"`
	Hosts  map[string]Define.HostConfig `yaml:"Hosts,omitempty"`
}

type YAMLOutput struct {
	Global  map[string]string      `yaml:"global,omitempty"`
	Default Define.HostConfig      `yaml:"default,omitempty"`
	Groups  map[string]GroupConfig `yaml:",inline"`
}

func ConvertToYAML(hostConfigs []Define.HostConfig) []byte {
	var output YAMLOutput

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
		groups := make(map[string]GroupConfig)

		for _, config := range normalConfigs {
			groupHostConfig := Define.HostConfig{}
			if config.Notes != "" {
				groupHostConfig.Notes = config.Notes
			}
			groupHostConfig.Config = config.Config

			groups[fmt.Sprintf("Group %s", config.Name)] = GroupConfig{
				Config: Define.HostConfig{},
				Hosts: map[string]Define.HostConfig{
					config.Name: groupHostConfig,
				},
			}
		}

		output.Groups = groups
	}

	return Fn.GetYamlBytes(output)
}
