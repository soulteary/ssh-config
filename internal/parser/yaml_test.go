package parser_test

import (
	"os"
	"path"
	"reflect"
	"testing"

	Define "github.com/soulteary/ssh-config/internal/define"
	Fn "github.com/soulteary/ssh-config/internal/fn"
	Parser "github.com/soulteary/ssh-config/internal/parser"
)

func TestConvertToYAML(t *testing.T) {
	tests := []struct {
		name     string
		input    []Define.HostConfig
		expected Define.YAMLOutput
	}{
		{
			name:     "Empty input",
			input:    []Define.HostConfig{},
			expected: Define.YAMLOutput{},
		},
		{
			name: "Only global config",
			input: []Define.HostConfig{
				{
					Name: "*",
					Config: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
			},
			expected: Define.YAMLOutput{
				Global: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
		},
		{
			name: "Only normal config",
			input: []Define.HostConfig{
				{
					Name:  "host1",
					Notes: "note1",
					Config: map[string]string{
						"key1": "value1",
					},
				},
			},
			expected: Define.YAMLOutput{
				Groups: map[string]Define.GroupConfig{
					"Group host1": {
						Common: make(map[string]string),
						Hosts: map[string]Define.HostConfig{
							"host1": {
								Notes: "note1",
								Config: map[string]string{
									"key1": "value1",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Both global and normal config",
			input: []Define.HostConfig{
				{
					Name: "*",
					Config: map[string]string{
						"globalKey": "globalValue",
					},
				},
				{
					Name:  "host1",
					Notes: "note1",
					Config: map[string]string{
						"key1": "value1",
					},
				},
			},
			expected: Define.YAMLOutput{
				Global: map[string]string{
					"globalKey": "globalValue",
				},
				Groups: map[string]Define.GroupConfig{
					"Group host1": {
						Common: make(map[string]string),
						Hosts: map[string]Define.HostConfig{
							"host1": {
								Notes: "note1",
								Config: map[string]string{
									"key1": "value1",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parser.ConvertToYAML(tt.input)
			if string(Fn.GetYamlBytes(tt.expected)) != string(result) {
				t.Errorf("ConvertToYAML() got = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFindGlobalConfig(t *testing.T) {
	input := []Define.HostConfig{
		{Name: "*", Config: map[string]string{"key": "value"}},
		{Name: "host1", Config: map[string]string{"key": "value"}},
	}
	expected := []Define.HostConfig{{Name: "*", Config: map[string]string{"key": "value"}}}

	result := Fn.FindGlobalConfig(input)
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("FindGlobalConfig() got = %v, want %v", result, expected)
	}
}

func TestFindNormalConfig(t *testing.T) {
	input := []Define.HostConfig{
		{Name: "*", Config: map[string]string{"key": "value"}},
		{Name: "host1", Config: map[string]string{"key": "value"}},
	}
	expected := []Define.HostConfig{{Name: "host1", Config: map[string]string{"key": "value"}}}

	result := Fn.FindNormalConfig(input)
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("FindNormalConfig() got = %v, want %v", result, expected)
	}
}

func TestGroupYAMLConfig(t *testing.T) {
	// Test case 1: Empty input
	input1 := ""
	result1 := Parser.GroupYAMLConfig(input1)
	if len(result1) != 0 {
		t.Errorf("Empty input should return empty result, got %v", result1)
	}

	// Test case 2: Only global config
	input2 := `
global:
  user: globaluser
  port: "22"
`
	expected2 := []Define.HostConfig{
		{
			Name: "*",
			Config: map[string]string{
				"user": "globaluser",
				"port": "22",
			},
		},
	}
	result2 := Parser.GroupYAMLConfig(input2)
	if !reflect.DeepEqual(expected2, result2) {
		t.Errorf("Global config not correctly parsed. Expected %v, got %v", expected2, result2)
	}

	pwd, err := os.Getwd()
	if err != nil {
		t.Errorf("TestGroupYAMLConfig() error = %v", err)
	}

	buf, err := os.ReadFile(path.Join(pwd, "../../testdata/parser-yaml-group.yaml"))
	if err != nil {
		t.Errorf("TestGroupYAMLConfig() error = %v", err)
	}

	// Test case 3: Global config and groups
	input3 := string(buf)
	expected3 := []Define.HostConfig{
		{
			Name: "*",
			Config: map[string]string{
				"HostKeyAlgorithms":        "+ssh-rsa",
				"PubkeyAcceptedAlgorithms": "+ssh-rsa",
			},
		},
		{
			Name:  "server1",
			Notes: "your notes here",
			Config: map[string]string{
				"Compression":    "yes",
				"ControlPath":    "~/.ssh/server-1-%r@%h:%p",
				"ControlPersist": "yes",
				"ForwardAgent":   "yes",
				"HostName":       "123.123.123.123",
				"IdentityFile":   "~/.ssh/keys/your-key1",
				"Port":           "1234",
				"TCPKeepAlive":   "yes",
			},
		},
		{
			Name: "server2",
			Config: map[string]string{
				"Compression":    "yes",
				"ControlPath":    "~/.ssh/server-2-%r@%h:%p",
				"ControlPersist": "yes",
				"ForwardAgent":   "yes",
				"HostName":       "123.234.123.234",
				"IdentityFile":   "~/.ssh/keys/your-key2",
				"Port":           "1234",
				"TCPKeepAlive":   "yes",
				"User":           "ubuntu",
			},
		},
	}
	result3 := Parser.GroupYAMLConfig(input3)

	if len(result3) != len(expected3) {
		t.Errorf("Global config and groups not correctly parsed. Expected %v, got %v", expected3, result3)
	}

	for resultItem := range result3 {
		resultLabel := result3[resultItem].Name
		for expectItem := range expected3 {
			if expected3[expectItem].Name == resultLabel {

				if expected3[expectItem].Notes != result3[resultItem].Notes {
					t.Errorf("Notes not correctly parsed. Expected %v, got %v", expected3[expectItem].Notes, result3[resultItem].Notes)
				}

				orderKeys := Fn.GetOrderMaps(expected3[expectItem].Config)
				for _, key := range orderKeys.Keys {
					if expected3[expectItem].Config[key] != result3[resultItem].Config[key] {
						t.Errorf("Config not correctly parsed. Expected %v, got %v", expected3[expectItem].Config[key], result3[resultItem].Config[key])
					}
				}
			}
		}
	}
}

func TestGroupYAMLConfigWithGroupPrefix(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Errorf("TestGroupYAMLConfigWithGroupPrefix() error = %v", err)
	}

	buf, err := os.ReadFile(path.Join(pwd, "../../testdata/parser-yaml-with-group-prefix.yaml"))
	if err != nil {
		t.Errorf("TestGroupYAMLConfigWithGroupPrefix() error = %v", err)
	}

	input := string(buf)
	expected := []Define.HostConfig{
		{
			Name: "*",
			Config: map[string]string{
				"HostKeyAlgorithms":        "+ssh-rsa",
				"PubkeyAcceptedAlgorithms": "+ssh-rsa",
			},
		},
		{
			Name:  "server1",
			Notes: "your notes here",
			Config: map[string]string{
				"Compression":    "yes",
				"ControlPath":    "~/.ssh/server-1-%r@%h:%p",
				"ControlPersist": "yes",
				"ForwardAgent":   "yes",
				"HostName":       "123.123.123.123",
				"IdentityFile":   "~/.ssh/keys/your-key1",
				"Port":           "1234",
				"TCPKeepAlive":   "yes",
			},
			Extra: Define.HostExtraConfig{
				Prefix: "public-",
			},
		},
		{
			Name: "server2",
			Config: map[string]string{
				"Compression":    "yes",
				"ControlPath":    "~/.ssh/server-2-%r@%h:%p",
				"ControlPersist": "yes",
				"ForwardAgent":   "yes",
				"HostName":       "123.234.123.234",
				"IdentityFile":   "~/.ssh/keys/your-key2",
				"Port":           "1234",
				"TCPKeepAlive":   "yes",
				"User":           "ubuntu",
			},
			Extra: Define.HostExtraConfig{
				Prefix: "public-",
			},
		},
	}
	result := Parser.GroupYAMLConfig(input)

	if len(result) != len(expected) {
		t.Errorf("Global config and groups not correctly parsed. Expected %v, got %v", expected, result)
	}

	for resultItem := range result {
		resultLabel := result[resultItem].Name
		for expectItem := range expected {
			if expected[expectItem].Name == resultLabel {

				if expected[expectItem].Notes != result[resultItem].Notes {
					t.Errorf("Notes not correctly parsed. Expected %v, got %v", expected[expectItem].Notes, result[resultItem].Notes)
				}

				orderKeys := Fn.GetOrderMaps(expected[expectItem].Config)
				for _, key := range orderKeys.Keys {
					if expected[expectItem].Config[key] != result[resultItem].Config[key] {
						t.Errorf("Config not correctly parsed. Expected %v, got %v", expected[expectItem].Config[key], result[resultItem].Config[key])
					}
				}
			}
		}
	}
}

func TestYAMLConfigWithDefault(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Errorf("TestYAMLConfigWithDefault() error = %v", err)
	}

	buf, err := os.ReadFile(path.Join(pwd, "../../testdata/parser-yaml-with-default.yaml"))
	if err != nil {
		t.Errorf("TestYAMLConfigWithDefault() error = %v", err)
	}

	input := string(buf)
	expected := []Define.HostConfig{
		{
			Name: "*",
			Config: map[string]string{
				"HostKeyAlgorithms":        "+ssh-rsa",
				"PubkeyAcceptedAlgorithms": "+ssh-rsa",
			},
		},
		{
			Name:  "server1",
			Notes: "your notes here",
			Config: map[string]string{
				"Compression":    "yes",
				"ControlPath":    "~/.ssh/server-1-%r@%h:%p",
				"ControlPersist": "yes",
				"ForwardAgent":   "yes",
				"HostName":       "123.123.123.123",
				"IdentityFile":   "~/.ssh/keys/your-key1",
				"Port":           "1234",
				"TCPKeepAlive":   "yes",
			},
			Extra: Define.HostExtraConfig{
				Prefix: "public-",
			},
		},
		{
			Name: "server2",
			Config: map[string]string{
				"Compression":    "yes",
				"ControlPath":    "~/.ssh/server-2-%r@%h:%p",
				"ControlPersist": "yes",
				"ForwardAgent":   "yes",
				"HostName":       "123.234.123.234",
				"IdentityFile":   "~/.ssh/keys/your-key2",
				"Port":           "1234",
				"TCPKeepAlive":   "yes",
				"User":           "ubuntu",
			},
			Extra: Define.HostExtraConfig{
				Prefix: "public-",
			},
		},
	}
	result := Parser.GroupYAMLConfig(input)

	if len(result) != len(expected) {
		t.Errorf("Global config and groups not correctly parsed. Expected %v, got %v", expected, result)
	}

	for resultItem := range result {
		resultLabel := result[resultItem].Name
		for expectItem := range expected {
			if expected[expectItem].Name == resultLabel {

				if expected[expectItem].Notes != result[resultItem].Notes {
					t.Errorf("Notes not correctly parsed. Expected %v, got %v", expected[expectItem].Notes, result[resultItem].Notes)
				}

				orderKeys := Fn.GetOrderMaps(expected[expectItem].Config)
				for _, key := range orderKeys.Keys {
					if expected[expectItem].Config[key] != result[resultItem].Config[key] {
						t.Errorf("Config not correctly parsed. Expected %v, got %v", expected[expectItem].Config[key], result[resultItem].Config[key])
					}
				}
			}
		}
	}
}
