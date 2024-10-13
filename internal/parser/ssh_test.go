package parser_test

import (
	"os"
	"path"
	"reflect"
	"testing"

	Parser "github.com/soulteary/ssh-yaml/internal/parser"
)

func TestGroupSSHConfig(t *testing.T) {

	pwd, err := os.Getwd()
	if err != nil {
		t.Errorf("TestGroupSSHConfig() error = %v", err)
	}

	buf, err := os.ReadFile(path.Join(pwd, "../../testdata/parser-ssh-group.cfg"))
	if err != nil {
		t.Errorf("TestGroupSSHConfig() error = %v", err)
	}

	actual := Parser.GroupSSHConfig(string(buf))
	if len(actual) != 3 {
		t.Errorf("TestGroupSSHConfig() = %v, want %v", len(actual), 3)
	}

	if _, ok := actual["server-cn-1"]; !ok {
		t.Errorf("TestGroupSSHConfig() = %v, want %v", ok, true)
	}

	if len(actual["server-cn-1"].Comments) == 0 {
		t.Errorf("TestGroupSSHConfig() = %v, want %v", len(actual["server-cn-1"].Comments), 1)
	}

	if _, ok := actual["server-us-2"]; !ok {
		t.Errorf("TestGroupSSHConfig() = %v, want %v", ok, true)
	}

	if len(actual["server-us-2"].Comments) == 0 {
		t.Errorf("TestGroupSSHConfig() = %v, want %v", len(actual["server-us-2"].Comments), 1)
	}

	if _, ok := actual["server-sg-3"]; !ok {
		t.Errorf("TestGroupSSHConfig() = %v, want %v", ok, true)
	}

	if len(actual["server-sg-3"].Comments) == 0 {
		t.Errorf("TestGroupSSHConfig() = %v, want %v", len(actual["server-sg-3"].Comments), 1)
	}
}

func TestGetSSHConfigContent(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		input    Parser.SSHHostConfigGroup
		expected Parser.SSHHostConfigGrouped
	}{
		{
			name: "Basic test",
			host: "example.com",
			input: Parser.SSHHostConfigGroup{
				Comments: []string{"# This is a comment"},
				Config: map[string]string{
					"HostName": "192.168.1.1",
					"User":     "admin",
					"Port":     "22",
				},
			},
			expected: Parser.SSHHostConfigGrouped{
				Comments: []string{"# This is a comment"},
				Config: "Host example.com\n" +
					"    HostName 192.168.1.1\n" +
					"    Port 22\n" +
					"    User admin",
			},
		},
		{
			name: "Empty config",
			host: "empty.host",
			input: Parser.SSHHostConfigGroup{
				Comments: []string{},
				Config:   map[string]string{},
			},
			expected: Parser.SSHHostConfigGrouped{
				Comments: []string{},
				Config:   "Host empty.host\n",
			},
		},
		{
			name: "Multiple comments",
			host: "multi.comment",
			input: Parser.SSHHostConfigGroup{
				Comments: []string{"# Comment 1", "# Comment 2"},
				Config: map[string]string{
					"Key": "Value",
				},
			},
			expected: Parser.SSHHostConfigGrouped{
				Comments: []string{"# Comment 1", "# Comment 2"},
				Config:   "Host multi.comment\n    Key Value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parser.GetSSHConfigContent(tt.host, tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("GetSSHConfigContent() = %v, want %v", result, tt.expected)
			}
		})
	}
}
