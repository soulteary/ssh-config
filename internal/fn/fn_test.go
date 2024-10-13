package fn_test

import (
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"

	Define "github.com/soulteary/ssh-yaml/internal/define"
	Fn "github.com/soulteary/ssh-yaml/internal/fn"
)

func TestGetUserInputFromStdin(t *testing.T) {
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Single line input",
			input:    "Hello, World!",
			expected: "Hello, World!",
		},
		{
			name:     "Multi-line input",
			input:    "Line 1\nLine 2\nLine 3",
			expected: "Line 1\nLine 2\nLine 3",
		},
		{
			name:     "Empty input",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)

			reader, writer, _ := os.Pipe()
			os.Stdin = reader

			go func() {
				io.Copy(writer, r)
				writer.Close()
			}()

			result := Fn.GetUserInputFromStdin()

			if result != tt.expected {
				t.Errorf("Expected %q, but got %q", tt.expected, result)
			}
		})
	}
}
func TestGetOrderMaps(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]string
		expected Fn.OrderedMap
	}{
		{
			name:  "Empty map",
			input: map[string]string{},
			expected: Fn.OrderedMap{
				Keys: []string{},
				Data: map[string]string{},
			},
		},
		{
			name: "Single item map",
			input: map[string]string{
				"key1": "value1",
			},
			expected: Fn.OrderedMap{
				Keys: []string{"key1"},
				Data: map[string]string{
					"key1": "value1",
				},
			},
		},
		{
			name: "Multiple items map",
			input: map[string]string{
				"key2": "value2",
				"key1": "value1",
				"key3": "value3",
			},
			expected: Fn.OrderedMap{
				Keys: []string{"key1", "key2", "key3"},
				Data: map[string]string{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Fn.GetOrderMaps(tt.input)

			// Check if Keys are correct and in order
			if !reflect.DeepEqual(result.Keys, tt.expected.Keys) {
				t.Errorf("Keys mismatch. Got %v, want %v", result.Keys, tt.expected.Keys)
			}

			// Check if Data map is correct
			if !reflect.DeepEqual(result.Data, tt.expected.Data) {
				t.Errorf("Data mismatch. Got %v, want %v", result.Data, tt.expected.Data)
			}
		})
	}
}
func TestGetYamlBytes(t *testing.T) {
	tests := []struct {
		name string
		data interface{}
		want []byte
	}{
		{
			name: "Simple struct",
			data: struct {
				Name string
				Age  int
			}{
				Name: "John Doe",
				Age:  30,
			},
			want: []byte("name: John Doe\nage: 30\n"),
		},
		{
			name: "Map",
			data: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			},
			want: []byte("key1: value1\nkey2: 42\n"),
		},
		{
			name: "Slice",
			data: []string{"apple", "banana", "cherry"},
			want: []byte("- apple\n- banana\n- cherry\n"),
		},
		{
			name: "Nil input",
			data: nil,
			want: []byte("null\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Fn.GetYamlBytes(tt.data)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetYamlBytes() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}

type UnmarshalableType struct{}

func (u UnmarshalableType) MarshalYAML() (interface{}, error) {
	return nil, fmt.Errorf("cannot marshal UnmarshalableType")
}

func TestGetYamlBytesError(t *testing.T) {
	t.Run("Invalid input", func(t *testing.T) {
		invalidData := UnmarshalableType{}
		result := Fn.GetYamlBytes(invalidData)
		if result != nil {
			t.Errorf("GetYamlBytes(%v) = %v, want nil", invalidData, result)
		}
	})
}

func TestGetJSONBytes(t *testing.T) {
	tests := []struct {
		name string
		data interface{}
		want []byte
	}{
		{
			name: "string",
			data: "test",
			want: []byte(`"test"`),
		},
		{
			name: "integer",
			data: 123,
			want: []byte(`123`),
		},
		{
			name: "float",
			data: 123.45,
			want: []byte(`123.45`),
		},
		{
			name: "boolean",
			data: true,
			want: []byte(`true`),
		},
		{
			name: "slice",
			data: []string{"a", "b", "c"},
			want: []byte(`["a","b","c"]`),
		},
		{
			name: "map",
			data: map[string]int{"a": 1, "b": 2},
			want: []byte(`{"a":1,"b":2}`),
		},
		{
			name: "struct",
			data: struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{"John", 30},
			want: []byte(`{"name":"John","age":30}`),
		},
		{
			name: "Nil input",
			data: nil,
			want: []byte(`null`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Fn.GetJSONBytes(tt.data)
			if string(got) != string(tt.want) {
				t.Errorf("GetJSONBytes() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}

type UnmarshalableJSONType struct{}

func (u UnmarshalableJSONType) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("cannot marshal UnmarshalableType")
}

func TestGetJSONBytes_Error(t *testing.T) {
	tests := []struct {
		name string
		data interface{}
		want []byte
	}{
		{
			name: "Function",
			data: func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Fn.GetJSONBytes(tt.data)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetJSONBytes() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}

func TestDetectStringType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Valid JSON",
			input:    `[{"HostName":"10.11.12.125"},{"HostName":"10.11.12.124"},{"HostName":"10.11.12.123"},{"Compression":"yes","ControlPath":"~/.ssh/server-%r@%h:%p","ControlPersist":"yes","ForwardAgent":"yes","HostName":"123.123.123.234","IdentityFile":"~/.ssh/keys/your-key","Port":"1234","TCPKeepAlive":"yes"},{"HostKeyAlgorithms":"+ssh-rsa"}]`,
			expected: "JSON",
		},
		{
			name: "Valid YAML",
			input: `
global:
  HostKeyAlgorithms: +ssh-rsa
Group 10.11.12.123:
  Hosts:
    10.11.12.124:
      config:
        HostName: 10.11.12.124
Group 10.11.12.125:
  Hosts:
    10.11.12.125:
      config:
        HostName: 10.11.12.125
Group server:
  Hosts:
    server:
      Notes: website
      config:
        Compression: "yes"
        ControlPath: ~/.ssh/server-%r@%h:%p
        ControlPersist: "yes"
        ForwardAgent: "yes"
        HostName: 123.123.123.234
        IdentityFile: ~/.ssh/keys/your-key
        Port: "1234"
        TCPKeepAlive: "yes"
			`,
			expected: "YAML",
		},
		{
			name:     "Plain text",
			input:    "This is just a plain text.",
			expected: "TEXT",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "TEXT",
		},
		{
			name:     "JSON-like Text",
			input:    "{name: test, value: 123}",
			expected: "TEXT",
		},
		{
			name:     "Whitespace-only input",
			input:    "   \n\t  ",
			expected: "TEXT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Fn.DetectStringType(tt.input)
			if result != tt.expected {
				t.Errorf("DetectStringType(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetYamlData(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Errorf("TestGetYamlData() error = %v", err)
	}

	buf, err := os.ReadFile(path.Join(pwd, "../../testdata/parser-yaml-group.yaml"))
	if err != nil {
		t.Errorf("TestGetYamlData() error = %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected Define.YAMLOutput
	}{
		{
			name:  "Valid YAML",
			input: string(buf),
			expected: Define.YAMLOutput{
				Global: map[string]string{
					"HostKeyAlgorithms":        "+ssh-rsa",
					"PubkeyAcceptedAlgorithms": "+ssh-rsa",
				},
				Groups: map[string]Define.GroupConfig{
					"Group public": {
						Hosts: map[string]Define.HostConfig{
							"server1": {
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
							"server2": {
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
						},
					},
				},
			},
		},
		{
			name:     "Empty input",
			input:    "",
			expected: Define.YAMLOutput{},
		},
		{
			name: "Invalid YAML",
			input: `
name: John Doe
age: thirty
`,
			expected: Define.YAMLOutput{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Fn.GetYamlData(tt.input)

			if result.Global != nil {
				for key, value := range result.Global {
					if tt.expected.Global[key] != value {
						t.Errorf("Global config mismatch. Expected %v, got %v", tt.expected.Global, result.Global)
					}
				}
			}

			if result.Groups != nil {
				for key, value := range result.Groups {
					for hostKey, hostValue := range value.Hosts {
						if tt.expected.Groups[key].Hosts[hostKey].Config != nil {
							for configKey, configValue := range hostValue.Config {
								if tt.expected.Groups[key].Hosts[hostKey].Config[configKey] != configValue {
									t.Errorf("Group config mismatch. Expected %v, got %v", tt.expected.Groups[key].Hosts[hostKey].Config, hostValue.Config)
								}
							}
						}
					}
				}
			}
		})
	}
}
