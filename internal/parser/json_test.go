/**
 * Copyright 2024-2025 Su Yang (soulteary)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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

func TestGroupJSONConfig(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []Define.HostConfig
	}{
		{
			name: "Normal case",
			input: `[
                {"Name": "host1", "Notes": "note1", "Data": {"key1": "value1", "key2": "value2"}},
                {"Name": "host2", "Notes": "note2", "Data": {"key3": "value3", "key4": "value4"}}
            ]`,
			expected: []Define.HostConfig{
				{
					Name:   "host1",
					Notes:  "note1",
					Config: map[string]string{"key1": "value1", "key2": "value2"},
				},
				{
					Name:   "host2",
					Notes:  "note2",
					Config: map[string]string{"key3": "value3", "key4": "value4"},
				},
			},
		},
		{
			name:     "Empty input",
			input:    "[]",
			expected: []Define.HostConfig{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Parser.GroupJSONConfig(tc.input)

			for i, hostConfig := range result {
				if hostConfig.Name != tc.expected[i].Name {
					t.Errorf("GroupJSONConfig(%s) = %v, want %v", tc.input, result, tc.expected)
				}
				if hostConfig.Notes != tc.expected[i].Notes {
					t.Errorf("GroupJSONConfig(%s) = %v, want %v", tc.input, result, tc.expected)
				}
				if !reflect.DeepEqual(hostConfig.Config, tc.expected[i].Config) {
					t.Errorf("GroupJSONConfig(%s) = %v, want %v", tc.input, result, tc.expected)
				}
			}

		})
	}
}
