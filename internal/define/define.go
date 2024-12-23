package define

type HostExtraConfig struct {
	Prefix string
}

// ssh config
type HostConfig struct {
	Name   string `yaml:"Name,omitempty"`
	Notes  string `yaml:"Notes,omitempty"`
	Config map[string]string
	Extra  HostExtraConfig `yaml:"Extra,omitempty"`
}

// json
type HostConfigDataForJSON map[string]string

type HostConfigForJSON struct {
	Name  string                `json:"Name,omitempty"`
	Notes string                `json:"Notes,omitempty"`
	Data  HostConfigDataForJSON `json:"Data,omitempty"`
}

// yaml
type GroupConfig struct {
	Prefix string                `yaml:"Prefix,omitempty"`
	Common map[string]string     `yaml:"Common,omitempty"`
	Hosts  map[string]HostConfig `yaml:"Hosts,omitempty"`
}
type YAMLOutput struct {
	Global  map[string]string      `yaml:"global,omitempty"`
	Default map[string]string      `yaml:"default,omitempty"`
	Groups  map[string]GroupConfig `yaml:",inline"`
}

var ExcludePatterns = []string{
	"known_hosts",
	"authorized_keys",
	"*.pub",
	"id_*",
	"*.key",
	"*.pem",
	"*.ppk",
}
