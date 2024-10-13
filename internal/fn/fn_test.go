package fn_test

import (
	"io"
	"os"
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
