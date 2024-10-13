package parser

type GroupConfig struct {
	Prefix string                `yaml:"Prefix,omitempty"`
	Config HostConfig            `yaml:"Config,omitempty"`
	Hosts  map[string]HostConfig `yaml:"Hosts,omitempty"`
}

type YAMLOutput struct {
	Global  map[string]string      `yaml:"global,omitempty"`
	Default HostConfig             `yaml:"default,omitempty"`
	Groups  map[string]GroupConfig `yaml:",inline"`
}
