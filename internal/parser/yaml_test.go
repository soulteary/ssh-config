package parser_test

import (
	"reflect"
	"testing"

	Define "github.com/soulteary/ssh-yaml/internal/define"
	Fn "github.com/soulteary/ssh-yaml/internal/fn"
	Parser "github.com/soulteary/ssh-yaml/internal/parser"
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
						Config: Define.HostConfig{},
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
						Config: Define.HostConfig{},
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
