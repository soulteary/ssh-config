package fn

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/soulteary/ssh-yaml/internal/define"
	Define "github.com/soulteary/ssh-yaml/internal/define"
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
