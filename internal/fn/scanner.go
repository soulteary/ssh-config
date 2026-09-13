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
	"io"
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

// ReadConfigFile reads one physical config without passing it through the
// legacy line scanner. This preserves long lines and every input byte for the
// default lossless conversion path.
func ReadConfigFile(path string) ([]byte, error) {
	file, err := openConfigFile(path)
	if err != nil {
		return nil, fmt.Errorf("open config file %s: %w", path, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("inspect config file %s: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("source path %s is not a regular file", path)
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read config file %s: %w", path, err)
	}
	return content, nil
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

// isLegacyConfigPath reports whether a path inside the scanned directory is one
// of the conventional OpenSSH client configuration locations.
//
// The scan used to accept any file holding one line that parsed as a known
// keyword, so an unrelated file kept under ~/.ssh was read in full and its
// comments were reproduced in the output as Notes. Deciding by path instead of
// by content removes that without making any judgement about what a file holds:
// a host stanza split across several files under config.d still resolves,
// because every one of those files is a configuration path whatever it
// contains.
func isLegacyConfigPath(root, path string) bool {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	relative = filepath.ToSlash(relative)
	return relative == "config" || strings.HasPrefix(relative, "config.d/")
}

// isLegacyDirectoryConfigFile reports whether a file found by the -legacy
// directory scan should be treated as configuration. When the scan is the
// implicit one over the user's ~/.ssh, files outside the conventional paths are
// skipped. For every file that is considered, the content test is the same as it
// has always been.
func isLegacyDirectoryConfigFile(root, path string, configPathsOnly bool) bool {
	if configPathsOnly && !isLegacyConfigPath(root, path) {
		return false
	}
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

// ReadSSHConfigs scans an explicitly named path. Every file is considered, as
// it always has been, because the caller chose the directory.
func ReadSSHConfigs(sshPath string) (*SSHConfig, error) {
	return readSSHConfigs(sshPath, false)
}

// ReadDefaultSSHConfigs scans the implicit ~/.ssh that -legacy falls back to
// when no source is given. There the directory's contents are not a deliberate
// choice, so only the conventional configuration paths are read.
func ReadDefaultSSHConfigs(sshPath string) (*SSHConfig, error) {
	return readSSHConfigs(sshPath, true)
}

func readSSHConfigs(sshPath string, configPathsOnly bool) (*SSHConfig, error) {
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

		if !isLegacyDirectoryConfigFile(sshPath, path, configPathsOnly) {
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
