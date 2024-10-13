package fn_test

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

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
		expected map[string]string
	}{
		{
			name:     "Empty map",
			input:    map[string]string{},
			expected: map[string]string{},
		},
		{
			name:     "Single key-value pair",
			input:    map[string]string{"a": "1"},
			expected: map[string]string{"a": "1"},
		},
		{
			name:     "Multiple key-value pairs",
			input:    map[string]string{"b": "2", "a": "1", "c": "3"},
			expected: map[string]string{"a": "1", "b": "2", "c": "3"},
		},
		{
			name:     "Keys with different cases",
			input:    map[string]string{"B": "2", "a": "1", "C": "3"},
			expected: map[string]string{"B": "2", "C": "3", "a": "1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Fn.GetOrderMaps(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("GetOrderMaps() = %v, want %v", result, tt.expected)
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
