package sshconfig

import (
	"strings"
	"testing"
)

func TestExpandEnvironmentOnlyAcceptsBracedNames(t *testing.T) {
	t.Parallel()
	environment := map[string]string{"PROFILE": "work", "EMPTY": ""}
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr string
	}{
		{name: "braced name", input: "hosts/${PROFILE}/config", want: "hosts/work/config"},
		{name: "unbraced dollar is literal", input: "hosts/$PROFILE/config", want: "hosts/$PROFILE/config"},
		{name: "double dollar is literal", input: "hosts/$$/config", want: "hosts/$$/config"},
		{name: "empty value", input: "hosts/${EMPTY}/config", want: "hosts//config"},
		{name: "missing name", input: "${MISSING}", wantErr: "is not set"},
		{name: "empty name", input: "${}", wantErr: "invalid environment variable name"},
		{name: "invalid name", input: "${BAD-NAME}", wantErr: "invalid environment variable name"},
		{name: "unterminated name", input: "${PROFILE", wantErr: "unterminated"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := expandEnvironment(test.input, environment)
			if test.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("expandEnvironment() error = %v, want %q", err, test.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Fatalf("expandEnvironment() = %q, want %q", got, test.want)
			}
		})
	}
}
