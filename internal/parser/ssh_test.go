package parser_test

import (
	"os"
	"path"
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
