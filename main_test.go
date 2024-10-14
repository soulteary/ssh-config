package main_test

import (
	"errors"
	"os"
	"path"
	"testing"

	Main "github.com/soulteary/ssh-config"
	Cmd "github.com/soulteary/ssh-config/cmd"
)

func TestRun(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Errorf("TestProcess() error = %v", err)
	}

	jsonContent, err := os.ReadFile(path.Join(pwd, "./testdata/main-test.json"))
	if err != nil {
		t.Errorf("TestProcess() error = %v", err)
	}

	sshContent, err := os.ReadFile(path.Join(pwd, "./testdata/main-test.cfg"))
	if err != nil {
		t.Errorf("TestProcess() error = %v", err)
	}

	yamlContent, err := os.ReadFile(path.Join(pwd, "./testdata/main-test.yaml"))
	if err != nil {
		t.Errorf("TestProcess() error = %v", err)
	}

	tests := []struct {
		name    string
		args    Cmd.Args
		deps    Main.Dependencies
		wantErr bool
	}{
		{
			name: "Invalid convert arguments",
			args: Cmd.Args{ToYAML: true, ToJSON: true, ToSSH: true},
			deps: Main.Dependencies{
				Println:       func(...interface{}) (int, error) { return 0, nil },
				CheckUseStdin: func() bool { return false },
			},
			wantErr: true,
		},
		{
			name: "Pipe mode",
			args: Cmd.Args{ToSSH: true},
			deps: Main.Dependencies{
				StdinStat:             func() (os.FileInfo, error) { return nil, nil },
				Println:               func(...interface{}) (int, error) { return 0, nil },
				GetUserInputFromStdin: func() string { return string(yamlContent) },
				Process:               func(string, string, Cmd.Args) []byte { return sshContent },
				CheckUseStdin:         func() bool { return true },
			},
			wantErr: false,
		},
		{
			name: "Invalid IO arguments",
			args: Cmd.Args{ToJSON: true, Src: "input.txt", Dest: "output.json"},
			deps: Main.Dependencies{
				StdinStat:     func() (os.FileInfo, error) { return nil, errors.New("not a pipe") },
				Println:       func(...interface{}) (int, error) { return 0, nil },
				CheckUseStdin: func() bool { return false },
			},
			wantErr: true,
		},
		{
			name: "File read error",
			args: Cmd.Args{ToJSON: true, Src: "input.txt", Dest: "output.json"},
			deps: Main.Dependencies{
				StdinStat:     func() (os.FileInfo, error) { return nil, errors.New("not a pipe") },
				Println:       func(...interface{}) (int, error) { return 0, nil },
				GetContent:    func(string) ([]byte, error) { return nil, errors.New("read error") },
				CheckUseStdin: func() bool { return false },
			},
			wantErr: true,
		},
		{
			name: "File save error",
			args: Cmd.Args{ToJSON: true, Src: "input.txt", Dest: "output.json"},
			deps: Main.Dependencies{
				StdinStat:     func() (os.FileInfo, error) { return nil, errors.New("not a pipe") },
				Println:       func(...interface{}) (int, error) { return 0, nil },
				GetContent:    func(string) ([]byte, error) { return sshContent, nil },
				SaveFile:      func(string, []byte) error { return errors.New("save error") },
				Process:       func(string, string, Cmd.Args) []byte { return jsonContent },
				CheckUseStdin: func() bool { return false },
			},
			wantErr: true,
		},
		{
			name: "Successful file conversion",
			args: Cmd.Args{ToYAML: true, Src: "testdata/main-test.cfg", Dest: "test.yaml"},
			deps: Main.Dependencies{
				StdinStat:     func() (os.FileInfo, error) { return nil, errors.New("not a pipe") },
				Println:       func(...interface{}) (int, error) { return 0, nil },
				GetContent:    func(string) ([]byte, error) { return sshContent, nil },
				SaveFile:      func(string, []byte) error { return nil },
				Process:       func(string, string, Cmd.Args) []byte { return yamlContent },
				CheckUseStdin: func() bool { return false },
			},
			wantErr: false,
		},
		{
			name: "File read error with print",
			args: Cmd.Args{ToJSON: true, Src: "testdata/main-test.cfg"},
			deps: Main.Dependencies{
				StdinStat:     func() (os.FileInfo, error) { return nil, errors.New("not a pipe") },
				Println:       func(...interface{}) (int, error) { return 0, nil },
				GetContent:    func(string) ([]byte, error) { return nil, errors.New("read error") },
				CheckUseStdin: func() bool { return false },
			},
			wantErr: true,
		},
		{
			name: "File save error with print",
			args: Cmd.Args{ToJSON: true, Src: "testdata/main-test.cfg"},
			deps: Main.Dependencies{
				StdinStat:     func() (os.FileInfo, error) { return nil, errors.New("not a pipe") },
				Println:       func(...interface{}) (int, error) { return 0, nil },
				GetContent:    func(string) ([]byte, error) { return sshContent, nil },
				SaveFile:      func(string, []byte) error { return errors.New("save error") },
				Process:       func(string, string, Cmd.Args) []byte { return jsonContent },
				CheckUseStdin: func() bool { return false },
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Main.Run(tt.args, tt.deps)
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
