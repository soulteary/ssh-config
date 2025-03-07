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

package fn_test

import (
	"reflect"
	"testing"

	Define "github.com/soulteary/ssh-config/internal/define"
	Fn "github.com/soulteary/ssh-config/internal/fn"
)

func TestFindGlobalConfig(t *testing.T) {
	configs := []Define.HostConfig{
		{Name: "host1", Notes: "Test host 1", Config: map[string]string{"key1": "value1"}},
		{Name: "*", Notes: "Global config", Config: map[string]string{"key2": "value2"}},
		{Name: "host2", Notes: "Test host 2", Config: map[string]string{"key3": "value3"}},
	}

	expected := Define.HostConfig{Name: "*", Notes: "Global config", Config: map[string]string{"key2": "value2"}}
	result := Fn.FindGlobalConfig(configs)

	if len(result) != 1 {
		t.Errorf("FindGlobalConfig() = %v, want %v", result, expected)
	}

	if !reflect.DeepEqual(result[0], expected) {
		t.Errorf("FindGlobalConfig() = %v, want %v", result, expected)
	}
}

func TestFindGlobalConfigs(t *testing.T) {
	configs := []Define.HostConfig{
		{Name: "host1", Notes: "Test host 1", Config: map[string]string{"key1": "value1"}},
		{Name: "*", Notes: "Global config", Config: map[string]string{"key2": "value2"}},
		{Name: "host2", Notes: "Test host 2", Config: map[string]string{"key3": "value3"}},
		{Name: "*", Notes: "Global config2", Config: map[string]string{"key3": "value4"}},
	}

	expected := []Define.HostConfig{
		{Name: "*", Notes: "Global config", Config: map[string]string{"key2": "value2"}},
		{Name: "*", Notes: "Global config2", Config: map[string]string{"key3": "value4"}},
	}
	result := Fn.FindGlobalConfig(configs)

	if len(result) != 2 {
		t.Errorf("FindGlobalConfigs() = %v, want %v", result, expected)
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("FindGlobalConfigs() = %v, want %v", result, expected)
	}
}

func TestFindGlobalConfigNoGlobal(t *testing.T) {
	configs := []Define.HostConfig{
		{Name: "host1", Notes: "Test host 1", Config: map[string]string{"key1": "value1"}},
		{Name: "host2", Notes: "Test host 2", Config: map[string]string{"key3": "value3"}},
	}

	result := Fn.FindGlobalConfig(configs)

	if len(result) != 0 {
		t.Errorf("FindGlobalConfig() = %v, want %v", result, 0)
	}
}

func TestFindNormalConfig(t *testing.T) {
	configs := []Define.HostConfig{
		{Name: "host1", Notes: "Test host 1", Config: map[string]string{"key1": "value1"}},
		{Name: "*", Notes: "Global config", Config: map[string]string{"key2": "value2"}},
		{Name: "host2", Notes: "Test host 2", Config: map[string]string{"key3": "value3"}},
	}

	expected := []Define.HostConfig{
		{Name: "host1", Notes: "Test host 1", Config: map[string]string{"key1": "value1"}},
		{Name: "host2", Notes: "Test host 2", Config: map[string]string{"key3": "value3"}},
	}
	result := Fn.FindNormalConfig(configs)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("FindNormalConfig() = %v, want %v", result, expected)
	}
}

func TestFindNormalConfigNoNormal(t *testing.T) {
	configs := []Define.HostConfig{
		{Name: "*", Notes: "Global config", Config: map[string]string{"key2": "value2"}},
	}

	result := Fn.FindNormalConfig(configs)

	if len(result) != 0 {
		t.Errorf("FindNormalConfig() = %v, want %v", result, 0)
	}
}
