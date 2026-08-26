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

	Cmd "github.com/soulteary/ssh-config/v2/cmd"
	Parser "github.com/soulteary/ssh-config/v2/internal/parser"
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
				Process:               func(string, string, Cmd.Args) ([]byte, error) { return sshContent, nil },
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
			name: "Process returns error",
			args: Cmd.Args{ToYAML: true, Src: "input.cfg", Dest: "out.yaml"},
			deps: Dependencies{
				StdinStat:     func() (os.FileInfo, error) { return nil, errors.New("not a pipe") },
				Println:       func(...interface{}) (int, error) { return 0, nil },
				GetContent:    func(string) ([]byte, error) { return []byte("Host \"unclosed"), nil },
				Process:       func(string, string, Cmd.Args) ([]byte, error) { return nil, errors.New("parsing config failed") },
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
				Process:       func(string, string, Cmd.Args) ([]byte, error) { return jsonContent, nil },
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
				Process:       func(string, string, Cmd.Args) ([]byte, error) { return yamlContent, nil },
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
				Process:       func(string, string, Cmd.Args) ([]byte, error) { return jsonContent, nil },
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
				Process:       func(string, string, Cmd.Args) ([]byte, error) { return jsonContent, nil },
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

func TestRunUsesAtomicSaverInLosslessMode(t *testing.T) {
	source := path.Join(t.TempDir(), "input.yaml")
	if err := os.WriteFile(source, []byte("schema"), 0600); err != nil {
		t.Fatal(err)
	}
	atomicCalled := false
	err := Run(Cmd.Args{ToSSH: true, Lossless: true, Src: source, Dest: "config"}, Dependencies{
		Println:       func(...interface{}) (int, error) { return 0, nil },
		CheckUseStdin: func() bool { return false },
		GetContent:    func(string) ([]byte, error) { return []byte("schema"), nil },
		Process:       func(string, string, Cmd.Args) ([]byte, error) { return []byte("Host example\n"), nil },
		SaveFile:      func(string, []byte) error { t.Fatal("legacy saver called"); return nil },
		SaveLossless: func(string, []byte) error {
			atomicCalled = true
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !atomicCalled {
		t.Fatal("lossless saver was not called")
	}
}

func TestRunLosslessPipePreservesInputAndOutputBytes(t *testing.T) {
	input := []byte("Host=example\r\nIdentityFile first\r\nIdentityFile second")
	var output []byte
	err := Run(Cmd.Args{ToSSH: true, Lossless: true}, Dependencies{
		Println:       func(...interface{}) (int, error) { return 0, nil },
		CheckUseStdin: func() bool { return true },
		GetUserInputFromStdin: func() string {
			t.Fatal("line-oriented stdin reader called")
			return ""
		},
		ReadStdin: func() ([]byte, error) { return input, nil },
		Process: func(_ string, got string, _ Cmd.Args) ([]byte, error) {
			if !bytes.Equal([]byte(got), input) {
				t.Fatalf("processor input = %q, want %q", got, input)
			}
			return []byte(got), nil
		},
		WriteOutput: func(data []byte) error {
			output = append([]byte(nil), data...)
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(output, input) {
		t.Fatalf("stdout = %q, want %q", output, input)
	}
}

func TestRunLosslessReadsItsStructuredOutput(t *testing.T) {
	original := []byte("Host=example\r\nIdentityFile first\r\nIdentityFile second")
	formats := []struct {
		name string
		args Cmd.Args
	}{
		{name: "YAML", args: Cmd.Args{ToYAML: true, Lossless: true}},
		{name: "JSON", args: Cmd.Args{ToJSON: true, Lossless: true}},
	}

	for _, format := range formats {
		t.Run(format.name, func(t *testing.T) {
			structured, err := Parser.Process("TEXT", string(original), format.args)
			if err != nil {
				t.Fatal(err)
			}
			var output []byte
			err = Run(Cmd.Args{ToSSH: true, Lossless: true}, Dependencies{
				Println:               func(...interface{}) (int, error) { return 0, nil },
				CheckUseStdin:         func() bool { return true },
				ReadStdin:             func() ([]byte, error) { return structured, nil },
				GetUserInputFromStdin: func() string { t.Fatal("legacy stdin reader called"); return "" },
				Process:               Parser.Process,
				WriteOutput: func(data []byte) error {
					output = append([]byte(nil), data...)
					return nil
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(output, original) {
				t.Fatalf("round trip mismatch\n got: %q\nwant: %q", output, original)
			}
		})
	}
}

func TestRunReportsStdinReadErrorsInLegacyMode(t *testing.T) {
	wantErr := errors.New("stdin failed")
	err := Run(Cmd.Args{ToYAML: true}, Dependencies{
		Println:       func(...interface{}) (int, error) { return 0, nil },
		CheckUseStdin: func() bool { return true },
		ReadStdin:     func() ([]byte, error) { return nil, wantErr },
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Run() error = %v, want %v", err, wantErr)
	}
}

func TestMainWithDependencies(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name           string
		args           []string
		expectedOutput string
		expectedError  string
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
			name:           "Successful execution to ssh",
			args:           []string{"cmd", "--to-ssh", "-src", "testdata/main-test.yaml", "-dest", "test.cfg"},
			expectedOutput: "File has been saved successfully\nFile path: test.cfg\n",
			expectedExit:   0,
			mockHomeDir:    os.UserHomeDir,
		},
		{
			name:          "Error execution",
			args:          []string{"cmd", "--to-json", "--to-yaml"}, // Invalid args
			expectedError: "Please specify either -to-yaml or -to-ssh or -to-json\n",
			expectedExit:  1,
			mockHomeDir:   os.UserHomeDir,
		},
		{
			name:          "Home directory error",
			args:          []string{"cmd"}, // No src specified, will try to use home dir
			expectedError: "Error: getting user home directory: mock home dir error\n",
			expectedExit:  1,
			mockHomeDir: func() (string, error) {
				return "", errors.New("mock home dir error")
			},
		},
		{
			name:           "Help exits before IO",
			args:           []string{"cmd", "-help"},
			expectedOutput: Cmd.Usage,
			expectedExit:   0,
			mockHomeDir: func() (string, error) {
				return "", errors.New("home must not be queried")
			},
		},
		{
			name:           "Version exits before IO",
			args:           []string{"cmd", "-version"},
			expectedOutput: versionText(),
			expectedExit:   0,
			mockHomeDir: func() (string, error) {
				return "", errors.New("home must not be queried")
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
			stdoutR, stdoutW, _ := os.Pipe()
			oldStderr := os.Stderr
			stderrR, stderrW, _ := os.Pipe()
			os.Stdout = stdoutW
			os.Stderr = stderrW

			MainWithDependencies(exitFunc, tt.mockHomeDir)
			Cmd.ResetFlags()

			stdoutW.Close()
			stderrW.Close()
			os.Stdout = oldStdout
			os.Stderr = oldStderr

			var stdout, stderr bytes.Buffer
			io.Copy(&stdout, stdoutR)
			io.Copy(&stderr, stderrR)
			output := stdout.String()

			if output != tt.expectedOutput {
				t.Errorf("Output = %q, want %q", output, tt.expectedOutput)
			}
			if got := stderr.String(); got != tt.expectedError {
				t.Errorf("Stderr = %q, want %q", got, tt.expectedError)
			}

			if exitCode != tt.expectedExit {
				t.Errorf("Exit code = %d, want %d", exitCode, tt.expectedExit)
			}

			if tt.expectedExit == 0 {
				for i := 0; i < len(tt.args)-1; i++ {
					if tt.args[i] == "-dest" || tt.args[i] == "--dest" {
						os.Remove(tt.args[i+1])
					}
				}
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
