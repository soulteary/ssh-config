package fn_test

import (
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	. "github.com/soulteary/ssh-yaml/internal/fn"
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

			result := GetUserInputFromStdin()

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
			result := GetOrderMaps(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("GetOrderMaps() = %v, want %v", result, tt.expected)
			}
		})
	}
}
