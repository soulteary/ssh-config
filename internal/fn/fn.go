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
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	Define "github.com/soulteary/ssh-config/v3/internal/define"
	"gopkg.in/yaml.v2"
	yamlv3 "gopkg.in/yaml.v3"
)

var (
	readFile  = os.ReadFile
	writeFile = os.WriteFile
	stat      = os.Stat
	mkdirAll  = os.MkdirAll
)

func GetUserInputFromStdin() string {
	input, _ := ReadUserInputFromStdin()
	return string(input)
}

// ReadUserInputFromStdin reads standard input without Scanner's token-size
// limit and returns any underlying read error to the caller.
func ReadUserInputFromStdin() ([]byte, error) {
	return io.ReadAll(os.Stdin)
}

type OrderedMap struct {
	Keys []string
	Data map[string]string
}

func GetOrderMaps(m map[string]string) OrderedMap {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	slices.Sort(keys)

	return OrderedMap{
		Keys: keys,
		Data: m,
	}
}

func GetYamlBytes(data any) []byte {
	yamlData, err := yaml.Marshal(&data)
	if err != nil {
		fmt.Println("Error marshaling to YAML:", err)
		return nil
	}
	return yamlData
}

func GetYamlData(input string) (yamlConfig Define.YAMLOutput) {
	yamlConfig, err := GetYamlDataStrict(input)
	if err != nil {
		fmt.Println("Error unmarshalling YAML:", err)
	}
	return yamlConfig
}

// GetYamlDataStrict decodes the legacy YAML format without hiding syntax or
// type errors from callers that must avoid destructive conversions.
func GetYamlDataStrict(input string) (yamlConfig Define.YAMLOutput, err error) {
	data := []byte(input)
	if err := validateLegacyYAML(data); err != nil {
		return yamlConfig, err
	}
	err = yamlv3.Unmarshal(data, &yamlConfig)
	return yamlConfig, err
}

func GetJSONBytes(data any) []byte {
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshaling to JSON:", err)
		return nil
	}
	return jsonData
}

func GetJSONData(input string) (jsonConfig []Define.HostConfigForJSON) {
	jsonConfig, err := GetJSONDataStrict(input)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
	}
	return jsonConfig
}

// GetJSONDataStrict decodes exactly one legacy JSON document and rejects
// unknown fields.
func GetJSONDataStrict(input string) (jsonConfig []Define.HostConfigForJSON, err error) {
	decoder := json.NewDecoder(strings.NewReader(input))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&jsonConfig); err != nil {
		return jsonConfig, err
	}

	var trailing json.RawMessage
	switch err := decoder.Decode(&trailing); err {
	case io.EOF:
		return jsonConfig, nil
	case nil:
		return jsonConfig, fmt.Errorf("unexpected JSON value after document")
	default:
		return jsonConfig, fmt.Errorf("invalid trailing JSON data: %w", err)
	}
}

func DetectStringType(input string) string {
	trimmedInput := strings.TrimSpace(input)

	if trimmedInput == "" {
		return "TEXT"
	}

	var js []Define.HostConfigForJSON
	if json.Unmarshal([]byte(trimmedInput), &js) == nil {
		return "JSON"
	}

	var y Define.YAMLOutput
	if yaml.Unmarshal([]byte(trimmedInput), &y) == nil {
		return "YAML"
	}
	return "TEXT"
}

// DetectStringTypeStrict preserves the legacy auto-detection behavior for
// valid input, but treats text that clearly starts as a structured document as
// malformed structured input instead of falling back to SSH text.
func DetectStringTypeStrict(input string) (string, error) {
	trimmedInput := strings.TrimSpace(input)
	if trimmedInput == "" {
		return "TEXT", nil
	}

	if strings.HasPrefix(trimmedInput, "{") || strings.HasPrefix(trimmedInput, "[") {
		var value []Define.HostConfigForJSON
		if err := json.Unmarshal([]byte(trimmedInput), &value); err != nil {
			return "", fmt.Errorf("invalid JSON input: %w", err)
		}
		return "JSON", nil
	}

	if looksLikeLegacyYAML(trimmedInput) {
		var value Define.YAMLOutput
		if err := yaml.Unmarshal([]byte(trimmedInput), &value); err != nil {
			return "", fmt.Errorf("invalid YAML input: %w", err)
		}
		return "YAML", nil
	}

	return DetectStringType(input), nil
}

func looksLikeLegacyYAML(input string) bool {
	for _, line := range strings.Split(input, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || line == "---" {
			continue
		}
		lower := strings.ToLower(line)
		return strings.HasPrefix(lower, "global:") ||
			strings.HasPrefix(lower, "default:") ||
			(strings.HasPrefix(lower, "group ") && strings.Contains(line, ":"))
	}
	return false
}

func GetPathContent(src string) ([]byte, error) {
	configFiles, err := ReadSSHConfigs(src)
	if err != nil {
		return nil, err
	}
	if len(configFiles.Configs) == 0 {
		return nil, fmt.Errorf("no valid SSH config found in %s", src)
	}

	filePaths := make([]string, 0, len(configFiles.Configs))
	for filePath := range configFiles.Configs {
		filePaths = append(filePaths, filePath)
	}
	slices.Sort(filePaths)

	var content []byte
	for _, filePath := range filePaths {
		fileContent, err := readFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("no valid SSH config found in %s: %w", src, err)
		}
		if len(content) > 0 && content[len(content)-1] != '\n' && content[len(content)-1] != '\r' {
			content = append(content, '\n')
		}
		content = append(content, fileContent...)
	}
	return content, nil
}

func Save(dest string, content []byte) error {
	destDir := filepath.Dir(dest)
	if err := ensureDirectory(destDir); err != nil {
		return err
	}

	info, err := stat(destDir)
	if err != nil {
		return fmt.Errorf("can not write to destination file: %v", err)
	}

	if !isDirWritable(info) {
		return fmt.Errorf("can not write to destination file: directory %s is not writable", destDir)
	}

	if err := writeFile(dest, content, 0644); err != nil {
		return fmt.Errorf("can not write to destination file: %v", err)
	}
	return nil
}

func TidyLastEmptyLines(input []byte) []byte {
	if len(input) == 0 {
		return input
	}

	end := len(input) - 1
	for end >= 0 && (input[end] == '\n' || input[end] == '\r') {
		end--
	}
	return input[:end+1]
}

func ensureDirectory(destDir string) error {
	info, err := stat(destDir)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("can not create destination directory: %s is not a directory", destDir)
		}
		return nil
	}

	if !os.IsNotExist(err) {
		return fmt.Errorf("can not create destination directory: %v", err)
	}

	parent := filepath.Dir(destDir)
	if parent != destDir {
		if parentInfo, parentErr := stat(parent); parentErr == nil {
			if !parentInfo.IsDir() {
				return fmt.Errorf("can not create destination directory: parent %s is not a directory", parent)
			}
			if !isDirWritable(parentInfo) {
				return fmt.Errorf("can not create destination directory: parent directory %s is not writable", parent)
			}
		}
	}

	if err := mkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("can not create destination directory: %v", err)
	}

	return nil
}
