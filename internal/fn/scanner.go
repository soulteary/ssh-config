package fn

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/soulteary/ssh-config/internal/define"
)

type ConfigFile struct {
	Path    string
	Content []string
	Hosts   map[string]map[string]string
}

type SSHConfig struct {
	Configs map[string]*ConfigFile // key: 配置文件路径
}

func IsExcluded(filename string) bool {
	filename = strings.ToLower(filename)

	for _, pattern := range define.ExcludePatterns {
		if matched, _ := filepath.Match(pattern, filename); matched {
			return true
		}
	}

	return false
}

func IsConfigFile(path string) bool {
	// read file first few lines to determine if it's SSH config file format
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	validLines := 0

	// check first 5 lines
	for scanner.Scan() && lineCount < 5 {
		line := strings.TrimSpace(scanner.Text())
		lineCount++

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			key := strings.ToLower(parts[0])
			switch key {
			case "host", "hostname", "user", "port", "identityfile", "proxycommand":
				validLines++
			}
		}
	}

	return validLines > 0
}

func ReadSSHConfigs(sshPath string) (*SSHConfig, error) {
	config := &SSHConfig{
		Configs: make(map[string]*ConfigFile),
	}

	info, err := os.Stat(sshPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get path object info: %v", err)
	}

	if !info.IsDir() {
		configFile := ReadSingleConfig(sshPath)
		if configFile != nil {
			config.Configs[sshPath] = configFile
		}
		return config, nil
	}

	err = filepath.Walk(sshPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if IsExcluded(info.Name()) {
			return nil
		}

		if !IsConfigFile(path) {
			return nil
		}

		configFile := ReadSingleConfig(path)
		if configFile != nil {
			config.Configs[path] = configFile
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %v", err)
	}

	return config, nil
}

func ReadSingleConfig(path string) *ConfigFile {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	config := &ConfigFile{
		Path:  path,
		Hosts: make(map[string]map[string]string),
	}

	scanner := bufio.NewScanner(file)
	var currentHost string
	var content []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		content = append(content, line)

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		key := strings.ToLower(parts[0])
		value := strings.Join(parts[1:], " ")

		if key == "host" {
			currentHost = value
			config.Hosts[currentHost] = make(map[string]string)
		} else if currentHost != "" {
			config.Hosts[currentHost][key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil
	}

	config.Content = content
	return config
}

func (c *SSHConfig) GetHostConfig(host string) map[string]map[string]string {
	results := make(map[string]map[string]string)

	for path, config := range c.Configs {
		if hostConfig, exists := config.Hosts[host]; exists {
			results[path] = hostConfig
		}
	}

	return results
}

func (c *SSHConfig) PrintConfigs() {
	for path, config := range c.Configs {
		fmt.Printf("\n=== 配置文件: %s ===\n", path)
		for host, hostConfig := range config.Hosts {
			fmt.Printf("\nHost %s:\n", host)
			for key, value := range hostConfig {
				fmt.Printf("  %s = %s\n", key, value)
			}
		}
	}
}
