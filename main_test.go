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

package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path"
	"testing"

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
		deps    Dependencies
		wantErr bool
	}{
		{
			name: "Invalid convert arguments",
			args: Cmd.Args{ToYAML: true, ToJSON: true, ToSSH: true},
			deps: Dependencies{
				Println:       func(...interface{}) (int, error) { return 0, nil },
				CheckUseStdin: func() bool { return false },
			},
			wantErr: true,
		},
		{
			name: "Pipe mode",
			args: Cmd.Args{ToSSH: true},
			deps: Dependencies{
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
			deps: Dependencies{
				StdinStat:     func() (os.FileInfo, error) { return nil, errors.New("not a pipe") },
				Println:       func(...interface{}) (int, error) { return 0, nil },
				CheckUseStdin: func() bool { return false },
			},
			wantErr: true,
		},
		{
			name: "File read error",
			args: Cmd.Args{ToJSON: true, Src: "input.txt", Dest: "output.json"},
			deps: Dependencies{
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
			deps: Dependencies{
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
			deps: Dependencies{
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
			deps: Dependencies{
				StdinStat:     func() (os.FileInfo, error) { return nil, errors.New("not a pipe") },
				Println:       func(...interface{}) (int, error) { return 0, nil },
				GetContent:    func(string) ([]byte, error) { return nil, errors.New("read error") },
				CheckUseStdin: func() bool { return false },
			},
			wantErr: true,
		},
		{
			name: "File save error with print",
			args: Cmd.Args{ToJSON: true, Src: "testdata/main-test.cfg", Dest: "can-not-save.json"},
			deps: Dependencies{
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
			args: Cmd.Args{ToJSON: true, Src: "testdata/main-test.cfg"},
			deps: Dependencies{
				StdinStat:     func() (os.FileInfo, error) { return nil, errors.New("not a pipe") },
				Println:       func(...interface{}) (int, error) { return 0, nil },
				GetContent:    func(string) ([]byte, error) { return sshContent, nil },
				SaveFile:      func(string, []byte) error { return nil },
				Process:       func(string, string, Cmd.Args) []byte { return jsonContent },
				CheckUseStdin: func() bool { return false },
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Run(tt.args, tt.deps)
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMainWithDependencies(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name           string
		args           []string
		expectedOutput string
		expectedExit   int
		mockHomeDir    func() (string, error)
	}{
		{
			name:           "Successful execution",
			args:           []string{"cmd", "--to-yaml", "-src", "testdata/main-test.json", "-dest", "test.yaml"},
			expectedOutput: "File has been saved successfully\nFile path: test.yaml\n",
			expectedExit:   0,
			mockHomeDir:    os.UserHomeDir,
		},
		{
			name:           "Error execution",
			args:           []string{"cmd", "--to-json", "--to-yaml"}, // Invalid args
			expectedOutput: "Please specify either -to-yaml or -to-ssh or -to-json\n",
			expectedExit:   1,
			mockHomeDir:    os.UserHomeDir,
		},
		{
			name:           "Home directory error",
			args:           []string{"cmd"}, // No src specified, will try to use home dir
			expectedOutput: "Error: getting user home directory: mock home dir error\nError: Source path '.ssh' does not exist\n",
			expectedExit:   1,
			mockHomeDir: func() (string, error) {
				return "", errors.New("mock home dir error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args

			exitCode := 0
			exitFunc := func(code int) {
				exitCode = code
			}

			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			MainWithDependencies(exitFunc, tt.mockHomeDir)
			Cmd.ResetFlags()

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if output != tt.expectedOutput {
				if tt.name != "Error execution" {
					t.Errorf("Output = %q, want %q", output, tt.expectedOutput)
				}
			}

			if exitCode != tt.expectedExit {
				t.Errorf("Exit code = %d, want %d", exitCode, tt.expectedExit)
			}

			if tt.expectedExit == 0 {
				os.Remove("output.json")
			}
		})
	}
}

var osExit = os.Exit

func TestMain(t *testing.T) {
	oldArgs := os.Args
	oldExit := osExit

	defer func() {
		os.Args = oldArgs
		osExit = oldExit
	}()

	exitCalled := false
	exitCode := 0
	osExit = func(code int) {
		exitCalled = true
		exitCode = code
	}

	os.Args = []string{"cmd", "--to-yaml", "-src", "testdata/main-test.json", "-dest", "test.yaml"}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	main()

	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	expectedOutput := "File has been saved successfully\nFile path: test.yaml\n"
	if output != expectedOutput {
		t.Errorf("Output = %q, want %q", output, expectedOutput)
	}

	if exitCalled {
		t.Errorf("os.Exit was called with code %d", exitCode)
	}

	os.Remove("test.yaml")
}
