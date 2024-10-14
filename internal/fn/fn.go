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
	srcInfo, err := os.Stat(src)
	if err != nil {
		return nil, fmt.Errorf("can not get source info: %v", err)
	}

	var content []byte

	if srcInfo.IsDir() {
		files, err := os.ReadDir(src)
		if err != nil {
			return nil, fmt.Errorf("can not read source directory: %v", err)
		}

		for _, file := range files {
			if !file.IsDir() {
				filePath := filepath.Join(src, file.Name())
				fileContent, err := os.ReadFile(filePath)
				if err != nil {
					return nil, fmt.Errorf("can not read file %s: %v", filePath, err)
				}
				content = append(content, fileContent...)
			}
		}
	} else {
		content, err = os.ReadFile(src)
		if err != nil {
			return nil, fmt.Errorf("can not read source file: %v", err)
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
