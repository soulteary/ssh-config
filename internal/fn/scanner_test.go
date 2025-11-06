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

package fn_test

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

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

	config := fn.ReadSingleConfig(configPath)
	if config == nil {
		t.Fatalf("readSingleConfig() returned nil")
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

func TestReadSSHConfigsErrors(t *testing.T) {
	// Test case 1: Non-existent path
	_, err := fn.ReadSSHConfigs("/non/existent/path")
	if err == nil {
		t.Error("Expected error for non-existent path, got nil")
	}

	// Test case 2: Single file with read error
	tmpfile, err := os.CreateTemp("", "ssh_config_test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Write some invalid content and change permissions to make it unreadable
	err = os.WriteFile(tmpfile.Name(), []byte("invalid content"), 0000)
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	_, err = fn.ReadSSHConfigs(tmpfile.Name())
	if err != nil {
		t.Error("Expected nil error for unreadable file(skip), got error")
	}

	// Test case 3: Directory with permission error
	dir, err := os.MkdirTemp("", "ssh_test_error")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(dir)

	// Create a subdirectory with no permissions
	noPermDir := filepath.Join(dir, "no_perm")
	if err := os.Mkdir(noPermDir, 0000); err != nil {
		t.Fatalf("Failed to create no-permission directory: %v", err)
	}

	// Try to read configs from the directory
	_, err = fn.ReadSSHConfigs(dir)
	if err == nil {
		t.Error("Expected error for directory with permission error, got nil")
	}

	// Test case 4: Invalid config file
	validDir, err := os.MkdirTemp("", "ssh_test_valid")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(validDir)

	invalidConfig := `\!@#$%^&
`
	err = os.WriteFile(filepath.Join(validDir, "config"), []byte(invalidConfig), 0600)
	if err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	// This should not return an error as we just log warnings for invalid configs
	config, err := fn.ReadSSHConfigs(validDir)
	if err != nil {
		t.Errorf("ReadSSHConfigs() unexpected error = %v", err)
	}
	if len(config.Configs) != 0 {
		t.Errorf("Expected 0 configs for invalid config file, got %d", len(config.Configs))
	}
}

func TestReadSSHConfigs_Walk_InaccessibleDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ssh-config-inaccessible")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	restrictedDir := filepath.Join(tempDir, "restricted")
	if err := os.Mkdir(restrictedDir, 0755); err != nil {
		t.Fatalf("Failed to create restricted directory: %v", err)
	}
	if err := os.Chmod(restrictedDir, 0300); err != nil {
		t.Fatalf("Failed to chmod restricted directory: %v", err)
	}
	defer os.Chmod(restrictedDir, 0755)

	_, err = fn.ReadSSHConfigs(tempDir)
	if err == nil {
		t.Fatal("Expected error for inaccessible directory, got nil")
	}
	if !strings.Contains(err.Error(), "not accessible") && !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("Expected not accessible or permission denied error, got: %v", err)
	}
}

func TestReadSSHConfigs_Walk_UnreadableFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ssh-config-unreadable")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config")
	content := "Host example\n  HostName example.com\n"
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	if err := os.Chmod(configPath, 0000); err != nil {
		t.Fatalf("Failed to chmod config file: %v", err)
	}

	if fn.IsConfigFile(configPath) {
		t.Log("config file detected as valid despite permissions")
	}

	cfg, err := fn.ReadSSHConfigs(tempDir)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(cfg.Configs) != 0 {
		t.Fatalf("Expected unreadable file to be skipped, got %d configs", len(cfg.Configs))
	}
}

func TestReadSSHConfigs_Walk_ErrorPropagation(t *testing.T) {
	attempts := 0
	for attempts < 5 {
		tempDir, err := os.MkdirTemp("", "ssh-config-walk-error")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}

		for i := 0; i < 200; i++ {
			subdir := filepath.Join(tempDir, fmt.Sprintf("dir_%03d", i))
			if err := os.MkdirAll(subdir, 0755); err != nil {
				t.Fatalf("Failed to create subdir: %v", err)
			}
			configPath := filepath.Join(subdir, "config")
			content := fmt.Sprintf("Host host-%d\n  HostName example.com\n", i)
			if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
				t.Fatalf("Failed to write config file: %v", err)
			}
		}

		errCh := make(chan struct{})
		go func(path string) {
			time.Sleep(5 * time.Millisecond)
			os.RemoveAll(path)
			close(errCh)
		}(tempDir)

		_, err = fn.ReadSSHConfigs(tempDir)
		<-errCh

		if err != nil {
			if strings.Contains(err.Error(), "failed to walk directory") {
				return
			}
		}

		os.RemoveAll(tempDir)
		attempts++
	}

	t.Fatal("expected error propagation from filepath.Walk")
}

func TestReadSSHConfigs_Walk(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "ssh-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test cases
	tests := []struct {
		name          string
		files         map[string]string // map of relative path to content
		expectedPaths []string          // expected config file paths
		expectedError bool
		setupFunc     func(dir string) error // optional setup function
	}{
		{
			name: "valid config files",
			files: map[string]string{
				"config":           "Host example\n  HostName example.com",
				"config.d/dev":     "Host dev\n  HostName dev.example.com",
				"config.d/staging": "Host staging\n  HostName staging.example.com",
			},
			expectedPaths: []string{
				filepath.Join(tempDir, "config"),
				filepath.Join(tempDir, "config.d/dev"),
				filepath.Join(tempDir, "config.d/staging"),
			},
		},
		{
			name: "mixed valid and invalid files",
			files: map[string]string{
				"config":           "Host example\n  HostName example.com",
				"config.d/dev":     "Host dev\n  HostName dev.example.com",
				".git/config":      "should be excluded",
				"readme.md":        "should be ignored",
				"config.d/.hidden": "should be excluded",
			},
			expectedPaths: []string{
				filepath.Join(tempDir, "config"),
				filepath.Join(tempDir, "config.d/dev"),
			},
		},
		{
			name: "invalid config content",
			files: map[string]string{
				"config":       "Host example\n  HostName example.com",
				"config.d/dev": "Invalid Config Content",
			},
			expectedPaths: []string{
				filepath.Join(tempDir, "config"),
			},
		},
		{
			name: "inaccessible directory",
			files: map[string]string{
				"config": "Host example\n  HostName example.com",
			},
			setupFunc: func(dir string) error {
				// Create a symbolic link to a non-existent directory
				return os.Symlink("/nonexistent", filepath.Join(dir, "broken-link"))
			},
			expectedPaths: []string{
				filepath.Join(tempDir, "config"),
			},
			expectedError: false, // We expect no error as we handle such errors gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up the temp directory
			if err := os.RemoveAll(tempDir); err != nil {
				t.Fatalf("Failed to clean temp dir: %v", err)
			}
			if err := os.MkdirAll(tempDir, 0755); err != nil {
				t.Fatalf("Failed to recreate temp dir: %v", err)
			}

			// Create test files
			for path, content := range tt.files {
				fullPath := filepath.Join(tempDir, path)
				dir := filepath.Dir(fullPath)
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatalf("Failed to create directory %s: %v", dir, err)
				}
				if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to write file %s: %v", fullPath, err)
				}
			}

			// Run optional setup
			if tt.setupFunc != nil {
				if err := tt.setupFunc(tempDir); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			// Run the function
			config, err := fn.ReadSSHConfigs(tempDir)

			// Check error
			if tt.expectedError {
				if err == nil {
					t.Error("Expected an error but got none")
				}
				return
			} else if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check number of configs
			if len(config.Configs) != len(tt.expectedPaths) {
				t.Errorf("Expected %d configs, got %d", len(tt.expectedPaths), len(config.Configs))
			}

			// Check if all expected paths are present
			for _, expectedPath := range tt.expectedPaths {
				if _, ok := config.Configs[expectedPath]; !ok {
					t.Errorf("Expected config for path %s not found", expectedPath)
				}
			}

			// Check if there are no unexpected paths
			for path := range config.Configs {
				found := false
				for _, expectedPath := range tt.expectedPaths {
					if path == expectedPath {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Found unexpected config for path: %s", path)
				}
			}
		})
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

// TestReadSingleConfig_ScannerError tests the scanner error handling
func TestReadSingleConfig_ScannerError(t *testing.T) {
	// 创建一个包含无效数据的文件
	tmpfile, err := os.CreateTemp("", "test_config_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// 写入一些正常数据和一个超长行来触发 scanner 错误
	longLine := strings.Repeat("a", bufio.MaxScanTokenSize*2)
	content := "Host testhost\n" + longLine

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// 测试 ReadSingleConfig
	result := fn.ReadSingleConfig(tmpfile.Name())

	// 验证当发生扫描错误时返回 nil
	if result != nil {
		t.Fatalf("Expected nil result when scanner error occurs, got: %v", result)
	}
}

func TestPrintConfigs(t *testing.T) {
	// 准备测试数据
	sshConfig := &fn.SSHConfig{
		Configs: map[string]*fn.ConfigFile{
			"/home/user/.ssh/config": {
				Path: "/home/user/.ssh/config",
				Content: []string{
					"Host github.com",
					"  HostName github.com",
					"  User git",
					"  IdentityFile ~/.ssh/github_rsa",
				},
				Hosts: map[string]map[string]string{
					"github.com": {
						"HostName":     "github.com",
						"User":         "git",
						"IdentityFile": "~/.ssh/github_rsa",
					},
				},
			},
			"/home/user/.ssh/config.d/work": {
				Path: "/home/user/.ssh/config.d/work",
				Content: []string{
					"Host work-server",
					"  HostName 192.168.1.100",
					"  User worker",
					"  Port 2222",
				},
				Hosts: map[string]map[string]string{
					"work-server": {
						"HostName": "192.168.1.100",
						"User":     "worker",
						"Port":     "2222",
					},
				},
			},
		},
	}

	// 捕获标准输出
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// 运行要测试的函数
	sshConfig.PrintConfigs()

	// 恢复标准输出并获取输出内容
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// 验证输出内容
	expectedOutputs := []string{
		"=== 配置文件: /home/user/.ssh/config ===",
		"Host github.com:",
		"  HostName = github.com",
		"  User = git",
		"  IdentityFile = ~/.ssh/github_rsa",
		"=== 配置文件: /home/user/.ssh/config.d/work ===",
		"Host work-server:",
		"  HostName = 192.168.1.100",
		"  User = worker",
		"  Port = 2222",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, but it didn't.\nActual output:\n%s", expected, output)
		}
	}
}
