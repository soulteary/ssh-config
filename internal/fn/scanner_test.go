package fn_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/soulteary/ssh-config/internal/fn"
)

func TestIsExcluded(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{"Known Hosts File", "known_hosts", true},
		{"Public Key", "id_rsa.pub", true},
		{"Private Key", "id_rsa", true},
		{"PEM File", "server.pem", true},
		{"PPK File", "key.ppk", true},
		{"Config File", "config", false},
		{"Custom Config", "ssh_config", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fn.IsExcluded(tt.filename); got != tt.want {
				t.Errorf("isExcluded() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createTempSSHConfig(t *testing.T, content string) (string, func()) {
	dir, err := os.MkdirTemp("", "ssh_config_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	configPath := filepath.Join(dir, "config")
	err = os.WriteFile(configPath, []byte(content), 0600)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatalf("Failed to write config file: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(dir)
	}

	return configPath, cleanup
}

func TestReadSingleConfig(t *testing.T) {
	configContent := `# SSH Config
Host github.com
    HostName github.com
    User git
    Port 22
    IdentityFile ~/.ssh/id_rsa

Host example
    HostName example.com
    User admin
    Port 2222
`

	configPath, cleanup := createTempSSHConfig(t, configContent)
	defer cleanup()

	config, err := fn.ReadSingleConfig(configPath)
	if err != nil {
		t.Fatalf("readSingleConfig() error = %v", err)
	}

	// Test parsed hosts
	expectedHosts := map[string]map[string]string{
		"github.com": {
			"hostname":     "github.com",
			"user":         "git",
			"port":         "22",
			"identityfile": "~/.ssh/id_rsa",
		},
		"example": {
			"hostname": "example.com",
			"user":     "admin",
			"port":     "2222",
		},
	}

	if !reflect.DeepEqual(config.Hosts, expectedHosts) {
		t.Errorf("readSingleConfig() hosts = %v, want %v", config.Hosts, expectedHosts)
	}
}

func TestReadSSHConfigs(t *testing.T) {
	// Create a temporary SSH directory with multiple config files
	dir, err := os.MkdirTemp("", "ssh_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(dir)

	// Create main config
	mainConfig := `Host main
    HostName main.example.com
    User mainuser
`
	err = os.WriteFile(filepath.Join(dir, "config"), []byte(mainConfig), 0600)
	if err != nil {
		t.Fatalf("Failed to write main config: %v", err)
	}

	// Create custom config
	customConfig := `Host custom
    HostName custom.example.com
    User customuser
`
	err = os.WriteFile(filepath.Join(dir, "custom_config"), []byte(customConfig), 0600)
	if err != nil {
		t.Fatalf("Failed to write custom config: %v", err)
	}

	// Create excluded files
	excluded := []string{
		"known_hosts",
		"id_rsa",
		"id_rsa.pub",
	}
	for _, filename := range excluded {
		err = os.WriteFile(filepath.Join(dir, filename), []byte("dummy content"), 0600)
		if err != nil {
			t.Fatalf("Failed to write excluded file %s: %v", filename, err)
		}
	}

	// Read configs
	sshConfig, err := fn.ReadSSHConfigs(dir)
	if err != nil {
		t.Fatalf("ReadSSHConfigs() error = %v", err)
	}

	// Check number of configs (should be 2 - main and custom)
	if len(sshConfig.Configs) != 2 {
		t.Errorf("ReadSSHConfigs() found %d configs, want 2", len(sshConfig.Configs))
	}

	// Test GetHostConfig
	mainHostConfig := sshConfig.GetHostConfig("main")
	if len(mainHostConfig) == 0 {
		t.Error("GetHostConfig() failed to find 'main' host")
	}

	customHostConfig := sshConfig.GetHostConfig("custom")
	if len(customHostConfig) == 0 {
		t.Error("GetHostConfig() failed to find 'custom' host")
	}
}

func TestIsConfigFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name: "Valid SSH Config",
			content: `Host example
    HostName example.com
    User test`,
			want: true,
		},
		{
			name: "Invalid File",
			content: `This is not a SSH config file
Just some random text
Nothing to see here`,
			want: false,
		},
		{
			name:    "Empty File",
			content: ``,
			want:    false,
		},
		{
			name: "Comments Only",
			content: `# This is a comment
# Another comment`,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, cleanup := createTempSSHConfig(t, tt.content)
			defer cleanup()

			if got := fn.IsConfigFile(path); got != tt.want {
				t.Errorf("isConfigFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
