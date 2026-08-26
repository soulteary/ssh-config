package fn_test

import (
	"testing"

	Fn "github.com/soulteary/ssh-config/v2/internal/fn"
)

func TestGetYamlDataStrictRejectsUnknownAndDuplicateFields(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"unknown group field": "group work:\n  Prefx: work-\n",
		"unknown host field":  "group work:\n  Hosts:\n    example:\n      Nmae: example\n",
		"duplicate field":     "group work:\n  Prefix: one\n  Prefix: two\n",
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
	if _, err := Fn.GetYamlDataStrict("group work:\n  Prefix: work-\n"); err != nil {
		t.Fatalf("GetYamlDataStrict() error = %v", err)
	}
	if _, err := Fn.GetJSONDataStrict("  [{\"Name\":\"example\",\"Data\":{\"User\":\"alice\"}}]\n"); err != nil {
		t.Fatalf("GetJSONDataStrict() error = %v", err)
	}
}
