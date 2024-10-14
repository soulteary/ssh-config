package fn_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	Define "github.com/soulteary/ssh-config/internal/define"
	Fn "github.com/soulteary/ssh-config/internal/fn"
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

func TestGetPathContent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	singleFile := filepath.Join(tempDir, "single.txt")
	singleContent := []byte("This is a single file")
	err = os.WriteFile(singleFile, singleContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create single file: %v", err)
	}

	content, err := Fn.GetPathContent(singleFile)
	if err != nil {
		t.Errorf("GetPathContent failed for single file: %v", err)
	}
	if !reflect.DeepEqual(content, singleContent) {
		t.Errorf("Expected %s, got %s", singleContent, content)
	}

	multiDir := filepath.Join(tempDir, "multi")
	err = os.Mkdir(multiDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create multi dir: %v", err)
	}

	file1 := filepath.Join(multiDir, "file1.txt")
	file2 := filepath.Join(multiDir, "file2.txt")
	content1 := []byte("Content of file 1")
	content2 := []byte("Content of file 2")

	err = os.WriteFile(file1, content1, 0644)
	if err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	err = os.WriteFile(file2, content2, 0644)
	if err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	content, err = Fn.GetPathContent(multiDir)
	if err != nil {
		t.Errorf("GetPathContent failed for directory: %v", err)
	}
	expectedContent := append(content1, content2...)
	if !reflect.DeepEqual(content, expectedContent) {
		t.Errorf("Expected %s, got %s", expectedContent, content)
	}

	nonExistentPath := filepath.Join(tempDir, "non_existent")
	_, err = Fn.GetPathContent(nonExistentPath)
	if err == nil {
		t.Error("Expected error for non-existent path, got nil")
	}

	unreadableDir := filepath.Join(tempDir, "unreadable")
	err = os.Mkdir(unreadableDir, 0000)
	if err != nil {
		t.Fatalf("Failed to create unreadable dir: %v", err)
	}
	defer os.Chmod(unreadableDir, 0755)

	_, err = Fn.GetPathContent(unreadableDir)
	if err == nil {
		t.Error("Expected error for unreadable directory, got nil")
	}

	dirWithUnreadableFile := filepath.Join(tempDir, "dir_with_unreadable")
	err = os.Mkdir(dirWithUnreadableFile, 0755)
	if err != nil {
		t.Fatalf("Failed to create dir_with_unreadable: %v", err)
	}

	unreadableFile := filepath.Join(dirWithUnreadableFile, "unreadable.txt")
	err = os.WriteFile(unreadableFile, []byte("Unreadable content"), 0000)
	if err != nil {
		t.Fatalf("Failed to create unreadable file: %v", err)
	}
	defer os.Chmod(unreadableFile, 0644)

	_, err = Fn.GetPathContent(dirWithUnreadableFile)
	if err == nil {
		t.Error("Expected error for directory with unreadable file, got nil")
	}

	unreadableFile2 := filepath.Join(tempDir, "unreadable_single.txt")
	err = os.WriteFile(unreadableFile2, []byte("Unreadable content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create unreadable single file: %v", err)
	}

	err = os.Chmod(unreadableFile2, 0000)
	if err != nil {
		t.Fatalf("Failed to change file permissions: %v", err)
	}
	defer os.Chmod(unreadableFile2, 0644)

	_, err = Fn.GetPathContent(unreadableFile2)
	if err == nil {
		t.Error("Expected error for unreadable single file, got nil")
	} else if !strings.Contains(err.Error(), "can not read source file") {
		t.Errorf("Expected error message to contain 'can not read source file', got: %v", err)
	}
}

func TestSave(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "test_save")
	defer os.RemoveAll(testDir)

	t.Run("successful save", func(t *testing.T) {
		dest := filepath.Join(testDir, "subdir", "test.txt")
		content := []byte("test content")

		err := Fn.Save(dest, content)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		savedContent, err := os.ReadFile(dest)
		if err != nil {
			t.Fatalf("Failed to read saved file: %v", err)
		}
		if !bytes.Equal(content, savedContent) {
			t.Fatalf("Saved content does not match. Expected %s, got %s", content, savedContent)
		}
	})

	t.Run("fail to create directory", func(t *testing.T) {
		readOnlyDir := filepath.Join(testDir, "readonly")
		err := os.MkdirAll(readOnlyDir, 0500)
		if err != nil {
			t.Fatalf("Failed to create read-only directory: %v", err)
		}

		dest := filepath.Join(readOnlyDir, "subdir", "test.txt")
		content := []byte("test content")

		err = Fn.Save(dest, content)
		if err == nil {
			t.Fatal("Expected an error, got nil")
		}
		if !strings.Contains(err.Error(), "can not create destination directory") {
			t.Fatalf("Unexpected error message: %v", err)
		}
	})

	t.Run("fail to write file", func(t *testing.T) {
		readOnlyDir := filepath.Join(testDir, "readonly_write")
		err := os.MkdirAll(readOnlyDir, 0500)
		if err != nil {
			t.Fatalf("Failed to create read-only directory: %v", err)
		}

		dest := filepath.Join(readOnlyDir, "test.txt")
		content := []byte("test content")

		err = Fn.Save(dest, content)
		if err == nil {
			t.Fatal("Expected an error, got nil")
		}
		if !strings.Contains(err.Error(), "can not write to destination file") {
			t.Fatalf("Unexpected error message: %v", err)
		}
	})
}

func TestGetJSONData(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Errorf("TestGetJSONData() error = %v", err)
	}

	buf, err := os.ReadFile(path.Join(pwd, "../../testdata/parser-json.json"))
	if err != nil {
		t.Errorf("TestGetJSONData() error = %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected []Define.HostConfigForJSON
	}{
		{
			name:  "Valid JSON",
			input: string(buf),
			expected: []Define.HostConfigForJSON{
				{
					Name: "*",
					Data: Define.HostConfigDataForJSON{
						"HostKeyAlgorithms":        "+ssh-rsa",
						"PubkeyAcceptedAlgorithms": "+ssh-rsa",
					},
				},
				{
					Name:  "server1",
					Notes: "your notes here",
					Data: Define.HostConfigDataForJSON{
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
					Data: Define.HostConfigDataForJSON{
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
		{
			name:     "Empty input",
			input:    "",
			expected: nil,
		},
		{
			name:     "Invalid JSON",
			input:    `{"invalid": "json"}`,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Fn.GetJSONData(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("GetJSONData() = %v, want %v", result, tt.expected)
			}
		})
	}
}
