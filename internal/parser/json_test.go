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
		expected []Define.HostConfigForJSON
	}{
		{
			name: "Single host config",
			input: []Define.HostConfig{
				{
					Name:  "example",
					Notes: "Test host",
					Config: map[string]string{
						"HostName": "example.com",
						"User":     "testuser",
						"Port":     "22",
					},
				},
			},
			expected: []Define.HostConfigForJSON{
				{
					Name:  "example",
					Notes: "Test host",
					Data: Define.HostConfigDataForJSON{
						"HostName": "example.com",
						"User":     "testuser",
						"Port":     "22",
					},
				},
			},
		},
		{
			name: "Multiple host configs",
			input: []Define.HostConfig{
				{
					Name:  "host1",
					Notes: "First host",
					Config: map[string]string{
						"HostName": "host1.com",
						"User":     "user1",
					},
				},
				{
					Name:  "host2",
					Notes: "Second host",
					Config: map[string]string{
						"HostName": "host2.com",
						"Port":     "2222",
					},
				},
			},
			expected: []Define.HostConfigForJSON{
				{
					Name:  "host1",
					Notes: "First host",
					Data: Define.HostConfigDataForJSON{
						"HostName": "host1.com",
						"User":     "user1",
					},
				},
				{
					Name:  "host2",
					Notes: "Second host",
					Data: Define.HostConfigDataForJSON{
						"HostName": "host2.com",
						"Port":     "2222",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Parser.ConvertToJSON(tc.input)

			var actualJSON []Define.HostConfigForJSON
			err := json.Unmarshal(result, &actualJSON)
			if err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			if !reflect.DeepEqual(actualJSON, tc.expected) {
				t.Errorf("ConvertToJSON() = %v, want %v", actualJSON, tc.expected)
			}
		})
	}
}
