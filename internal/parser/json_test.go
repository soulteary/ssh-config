package parser_test

import (
	"encoding/json"
	"reflect"
	"testing"

	Define "github.com/soulteary/ssh-config/internal/define"
	Parser "github.com/soulteary/ssh-config/internal/parser"
)

func TestConvertToJSON(t *testing.T) {
	testCases := []struct {
		name     string
		input    []Define.HostConfig
		expected []byte
	}{
		{
			name:     "Empty slice",
			input:    []Define.HostConfig{},
			expected: []byte("[]"),
		},
		{
			name: "Single host config",
			input: []Define.HostConfig{
				{
					Config: map[string]string{
						"Host":     "example.com",
						"Port":     "22",
						"Username": "user",
						"Password": "pass",
					},
				},
			},
			expected: []byte(`[{"Host":"example.com","Password":"pass","Port":"22","Username":"user"}]`),
		},
		{
			name: "Multiple host configs",
			input: []Define.HostConfig{
				{
					Config: map[string]string{
						"Host":     "example1.com",
						"Port":     "22",
						"Username": "user1",
						"Password": "pass1",
					},
				},
				{
					Config: map[string]string{
						"Host":     "example2.com",
						"Port":     "2222",
						"Username": "user2",
						"Password": "pass2",
					},
				},
			},
			expected: []byte(`[{"Host":"example1.com","Password":"pass1","Port":"22","Username":"user1"},{"Host":"example2.com","Password":"pass2","Port":"2222","Username":"user2"}]`),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Parser.ConvertToJSON(tc.input)

			var js interface{}
			err := json.Unmarshal(result, &js)
			if err != nil {
				t.Errorf("ConvertToJSON produced invalid JSON: %v", err)
			}

			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("ConvertToJSON() = %s, want %s", result, tc.expected)
			}
		})
	}
}
