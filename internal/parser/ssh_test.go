package parser_test

import (
	"os"
	"path"
	"reflect"
	"strings"
	"testing"

	Define "github.com/soulteary/ssh-yaml/internal/define"
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
				Comments: strings.Join([]string{"# This is a comment"}, "\n"),
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
				Comments: "",
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
				Comments: strings.Join([]string{"# Comment 1", "# Comment 2"}, "\n"),
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

func TestParseSSHConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		notes    string
		expected Parser.HostConfig
	}{
		{
			name: "Basic Config",
			input: `
Host example
    HostName example.com
    User testuser
    IdentityFile ~/.ssh/id_rsa
    Port 2222
	PubkeyAcceptedAlgorithms +ssh-rsa
`,
			notes: "",
			expected: Parser.HostConfig{
				HostName:                 "example.com",
				User:                     "testuser",
				IdentityFile:             "~/.ssh/id_rsa",
				Port:                     "2222",
				YamlUserHost:             "example",
				YamlUserNotes:            "",
				PubkeyAcceptedAlgorithms: "+ssh-rsa",
			},
		},
		{
			name: "Full Config",
			input: `
# This is a comment
Host fullexample
    # This is another comment
    HostName fullexample.com
    User fulluser
    IdentityFile ~/.ssh/full_id_rsa
    Port 3333
    ControlPath ~/.ssh/cm_%r@%h:%p
    ControlPersist 30m
    TCPKeepAlive yes
    Compression yes
    ForwardAgent yes
    Ciphers aes128-ctr,aes192-ctr,aes256-ctr
    HostKeyAlgorithms ssh-ed25519,rsa-sha2-512
    KexAlgorithms curve25519-sha256,diffie-hellman-group14-sha256
    PubkeyAuthentication yes
    ProxyCommand ssh jumphost nc %h %p
`,
			notes: strings.Join([]string{"# This is a comment", "# This is another comment", ""}, "\n"),
			expected: Parser.HostConfig{
				HostName:             "fullexample.com",
				User:                 "fulluser",
				IdentityFile:         "~/.ssh/full_id_rsa",
				Port:                 "3333",
				ControlPath:          "~/.ssh/cm_%r@%h:%p",
				ControlPersist:       "30m",
				TCPKeepAlive:         "yes",
				Compression:          "yes",
				ForwardAgent:         "yes",
				Ciphers:              "aes128-ctr,aes192-ctr,aes256-ctr",
				HostKeyAlgorithms:    "ssh-ed25519,rsa-sha2-512",
				KexAlgorithms:        "curve25519-sha256,diffie-hellman-group14-sha256",
				PubkeyAuthentication: "yes",
				ProxyCommand:         "ssh jumphost nc %h %p",
				YamlUserHost:         "fullexample",
				YamlUserNotes:        "# This is a comment\n# This is another comment\n",
			},
		},
		{
			name: "Empty Config",
			input: `
# This is a comment
Host empty
    # This is another comment
`,
			notes: strings.Join([]string{"# This is a comment", "# This is another comment", ""}, "\n"),
			expected: Parser.HostConfig{
				YamlUserHost:  "empty",
				YamlUserNotes: strings.Join([]string{"# This is a comment", "# This is another comment", ""}, "\n"),
			},
		},
		{
			name: "Unknown Keys",
			input: `
Host unknown
    HostName unknown.com
    UnknownKey1 value1
    UnknownKey2 value2
`,
			expected: Parser.HostConfig{
				HostName:     "unknown.com",
				YamlUserHost: "unknown",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Parser.ParseSSHConfig(tt.input, tt.notes)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("ParseSSHConfig() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetSingleHostData(t *testing.T) {
	tests := []struct {
		name           string
		input          Parser.HostConfig
		expectedResult map[string]string
		expectedName   string
		expectedNotes  string
	}{
		{
			name: "Full config",
			input: Parser.HostConfig{
				YamlUserHost:  "host1",
				YamlUserNotes: "note1",
				Port:          "22",
				User:          "user1",
			},
			expectedResult: map[string]string{
				"Port": "22",
				"User": "user1",
			},
			expectedName:  "host1",
			expectedNotes: "note1",
		},
		{
			name: "Partial config",
			input: Parser.HostConfig{
				YamlUserHost: "host2",
				Port:         "2222",
			},
			expectedResult: map[string]string{
				"Port": "2222",
			},
			expectedName:  "host2",
			expectedNotes: "",
		},
		{
			name:           "Empty config",
			input:          Parser.HostConfig{},
			expectedResult: map[string]string{},
			expectedName:   "",
			expectedNotes:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, name, notes := Parser.GetSingleHostData(tt.input)

			if !reflect.DeepEqual(result, tt.expectedResult) {
				t.Errorf("GetSingleHostData() result = %v, want %v", result, tt.expectedResult)
			}

			if name != tt.expectedName {
				t.Errorf("GetSingleHostData() name = %v, want %v", name, tt.expectedName)
			}

			if notes != tt.expectedNotes {
				t.Errorf("GetSingleHostData() notes = %v, want %v", notes, tt.expectedNotes)
			}
		})
	}
}

func TestConvertToSSH(t *testing.T) {
	tests := []struct {
		name     string
		input    []Define.HostConfig
		expected string
	}{
		{
			name:     "Empty input",
			input:    []Define.HostConfig{},
			expected: "",
		},
		{
			name: "Only global config",
			input: []Define.HostConfig{
				{
					Name:   "*",
					Notes:  "Global config",
					Config: map[string]string{"User": "globaluser", "IdentityFile": "~/.ssh/id_rsa"},
				},
			},
			expected: "# Global config\nHost *\n    IdentityFile ~/.ssh/id_rsa\n    User globaluser\n",
		},
		{
			name: "Only normal config",
			input: []Define.HostConfig{
				{
					Name:   "myserver",
					Notes:  "My server",
					Config: map[string]string{"HostName": "192.168.1.100", "User": "myuser"},
				},
			},
			expected: "# My server\nHost myserver\n    HostName 192.168.1.100\n    User myuser\n\n",
		},
		{
			name: "Both global and normal configs",
			input: []Define.HostConfig{
				{
					Name:   "*",
					Notes:  "Global config",
					Config: map[string]string{"User": "globaluser"},
				},
				{
					Name:   "myserver",
					Notes:  "My server",
					Config: map[string]string{"HostName": "192.168.1.100"},
				},
			},
			expected: "# Global config\nHost *\n    User globaluser\n\n# My server\nHost myserver\n    HostName 192.168.1.100\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parser.ConvertToSSH(tt.input)
			if !reflect.DeepEqual(string(result), tt.expected) {
				t.Errorf("ConvertToSSH() = \n```\n%v\n```\n, want ```\n%v\n```\n", string(result), tt.expected)
			}
		})
	}
}
