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

package fn

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/soulteary/ssh-config/v3/internal/define"
	"github.com/soulteary/ssh-config/v3/pkg/sshconfig"
)

const maxScannedConfigLine = 1024 * 1024

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
	return hasConfigDirective(path, func(string) bool { return true })
}

func isLegacyDirectoryConfigFile(path string) bool {
	return hasConfigDirective(path, func(keyword string) bool {
		return keyword != "include"
	})
}

func hasConfigDirective(path string, accept func(string) bool) bool {
	file, err := openConfigFile(path)
	if err != nil {
		return false
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return false
	}

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), maxScannedConfigLine)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		doc, err := sshconfig.Parse([]byte(line))
		if err != nil {
			continue
		}
		nodes := doc.Nodes()
		if len(nodes) == 1 && nodes[0].Directive != nil {
			keyword := nodes[0].Directive.KeywordValue
			if _, known := sshconfig.LookupKeyword(keyword); known && accept(keyword) {
				return true
			}
		}
	}
	return false
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
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("source path %s is not a regular file", sshPath)
		}
		if !isFileReadable(info) {
			return config, nil
		}

		configFile := ReadSingleConfig(sshPath)
		if configFile != nil {
			config.Configs[sshPath] = configFile
		}
		return config, nil
	}

	if !isDirReadable(info) {
		return nil, fmt.Errorf("failed to walk directory: %s is not accessible", sshPath)
	}

	err = filepath.Walk(sshPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if !isDirReadable(info) {
				return fmt.Errorf("directory %s is not accessible", path)
			}
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		if !info.Mode().IsRegular() {
			return nil
		}

		if IsExcluded(info.Name()) {
			return nil
		}

		if !isLegacyDirectoryConfigFile(path) {
			return nil
		}

		if !isFileReadable(info) {
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
	file, err := openConfigFile(path)
	if err != nil {
		return nil
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return nil
	}

	config := &ConfigFile{
		Path:  path,
		Hosts: make(map[string]map[string]string),
	}

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), maxScannedConfigLine)
	var currentHost string
	var content []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		content = append(content, line)

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 2 {
			key := strings.ToLower(parts[0])
			value := strings.Join(parts[1:], " ")

			if key == "host" {
				currentHost = value
				config.Hosts[currentHost] = make(map[string]string)
			} else if currentHost != "" {
				config.Hosts[currentHost][key] = value
			}
		}
	}

	scanErr := scanner.Err()
	if scanErr != nil {
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
