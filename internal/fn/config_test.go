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
