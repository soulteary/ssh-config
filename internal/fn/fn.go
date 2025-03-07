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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/soulteary/ssh-config/internal/define"
	Define "github.com/soulteary/ssh-config/internal/define"
	"gopkg.in/yaml.v2"
)

func GetUserInputFromStdin() string {
	var lines []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	input := strings.Join(lines, "\n")
	return input
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
	err := yaml.Unmarshal([]byte(input), &yamlConfig)
	if err != nil {
		fmt.Println("Error unmarshalling YAML:", err)
		return yamlConfig
	}
	return yamlConfig
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
	err := json.Unmarshal([]byte(input), &jsonConfig)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return jsonConfig
	}
	return jsonConfig
}

func DetectStringType(input string) string {
	trimmedInput := strings.TrimSpace(input)

	if trimmedInput == "" {
		return "TEXT"
	}

	var js []define.HostConfigForJSON
	if json.Unmarshal([]byte(trimmedInput), &js) == nil {
		return "JSON"
	}

	var y define.YAMLOutput
	if yaml.Unmarshal([]byte(trimmedInput), &y) == nil {
		return "YAML"
	}
	return "TEXT"
}

func GetPathContent(src string) ([]byte, error) {
	configFiles, err := ReadSSHConfigs(src)
	if err != nil {
		return nil, err
	}
	if len(configFiles.Configs) == 0 {
		return nil, fmt.Errorf("no valid SSH config found in %s", src)
	}

	var content []byte
	for filePath := range configFiles.Configs {
		fileContent, err := os.ReadFile(filePath)
		if err == nil {
			content = append(content, fileContent...)
		}
	}
	return content, nil
}

func Save(dest string, content []byte) error {
	destDir := filepath.Dir(dest)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("can not create destination directory: %v", err)
	}

	if err := os.WriteFile(dest, content, 0644); err != nil {
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
