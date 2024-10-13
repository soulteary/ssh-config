package define

type HostConfig struct {
	Name   string `yaml:"Name,omitempty"`
	Notes  string `yaml:"Notes,omitempty"`
	Config map[string]string
}

type HostConfigForJSON map[string]string
