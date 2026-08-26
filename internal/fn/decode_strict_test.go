package fn_test

import (
	"testing"

	Fn "github.com/soulteary/ssh-config/v2/internal/fn"
)

func TestGetYamlDataStrictRejectsUnknownAndDuplicateFields(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"unknown group field":        "group work:\n  Prefx: work-\n",
		"unknown merged group field": "group work:\n  <<: &defaults\n    Prefx: work-\n  Prefix: safe-\n",
		"unknown host field":         "group work:\n  Hosts:\n    example:\n      Nmae: example\n",
		"duplicate field":            "group work:\n  Prefix: one\n  Prefix: two\n",
	}
	for name, input := range tests {
		name, input := name, input
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := Fn.GetYamlDataStrict(input); err == nil {
				t.Fatal("GetYamlDataStrict() accepted an ambiguous legacy document")
			}
		})
	}
}

func TestGetJSONDataStrictRejectsUnknownAndTrailingValues(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"unknown field":  `[{"Name":"example","Datta":{"User":"alice"}}]`,
		"second value":   `[] []`,
		"trailing token": `[] invalid`,
	}
	for name, input := range tests {
		name, input := name, input
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := Fn.GetJSONDataStrict(input); err == nil {
				t.Fatal("GetJSONDataStrict() accepted an ambiguous legacy document")
			}
		})
	}
}

func TestStrictLegacyDecodersKeepValidInput(t *testing.T) {
	t.Parallel()
	yamlInput := "global:\n  <<: &defaults\n    User: alice\n    Port: \"22\"\n  User: bob\ngroup work:\n  Prefix: work-\n  Hosts:\n    example:\n      Extra:\n        prefix: host-\n"
	yamlConfig, err := Fn.GetYamlDataStrict(yamlInput)
	if err != nil {
		t.Fatalf("GetYamlDataStrict() error = %v", err)
	}
	if yamlConfig.Global["User"] != "bob" || yamlConfig.Global["Port"] != "22" {
		t.Fatalf("YAML merge override = %#v", yamlConfig.Global)
	}
	if prefix := yamlConfig.Groups["group work"].Hosts["example"].Extra.Prefix; prefix != "host-" {
		t.Fatalf("host prefix = %q, want host-", prefix)
	}
	overrideFirst := "global:\n  User: bob\n  <<: {User: alice, Port: \"22\"}\n"
	yamlConfig, err = Fn.GetYamlDataStrict(overrideFirst)
	if err != nil {
		t.Fatalf("GetYamlDataStrict() override-first error = %v", err)
	}
	if yamlConfig.Global["User"] != "bob" || yamlConfig.Global["Port"] != "22" {
		t.Fatalf("override-first YAML merge = %#v", yamlConfig.Global)
	}
	if _, err := Fn.GetJSONDataStrict("  [{\"Name\":\"example\",\"Data\":{\"User\":\"alice\"}}]\n"); err != nil {
		t.Fatalf("GetJSONDataStrict() error = %v", err)
	}
}
