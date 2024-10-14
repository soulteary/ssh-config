package main

import (
	"io"
	"os"
	"path"
	"strings"
	"testing"

	Cmd "github.com/soulteary/ssh-config/cmd"
	Fn "github.com/soulteary/ssh-config/internal/fn"
)

func TestMain(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	yamlContent, err := os.ReadFile(path.Join(pwd, "./testdata/main-test.yaml"))
	if err != nil {
		t.Fatalf("Failed to read YAML test file: %v", err)
	}

	sshContent, err := os.ReadFile(path.Join(pwd, "./testdata/main-test.cfg"))
	if err != nil {
		t.Fatalf("Failed to read SSH config test file: %v", err)
	}

	oldArgs := os.Args
	oldStdout := os.Stdout
	oldStdin := os.Stdin
	defer func() {
		os.Args = oldArgs
		os.Stdout = oldStdout
		os.Stdin = oldStdin
	}()

	const (
		TEST_INPUT_FILE  = "test_input"
		TEST_OUTPUT_FILE = "test_output"
	)

	tests := []struct {
		name     string
		args     []string
		input    string
		expected []byte
		isPipe   bool
		wantErr  bool
	}{
		{
			name:     "YAML file input to SSH file output",
			args:     []string{"cmd", "-src", TEST_INPUT_FILE + ".yaml", "-dest", TEST_OUTPUT_FILE + ".cfg", "-to-ssh"},
			input:    string(yamlContent),
			expected: sshContent,
			isPipe:   false,
			wantErr:  false,
		},
		{
			name:     "SSH file input to YAML file output",
			args:     []string{"cmd", "-src", TEST_INPUT_FILE + ".cfg", "-dest", TEST_OUTPUT_FILE + ".yaml", "-to-yaml"},
			input:    string(sshContent),
			expected: yamlContent,
			isPipe:   false,
			wantErr:  false,
		},
		{
			name:     "YAML pipe input to SSH stdout",
			args:     []string{"cmd", "-to-ssh"},
			input:    string(yamlContent),
			expected: sshContent,
			isPipe:   true,
			wantErr:  false,
		},
		{
			name:     "SSH pipe input to YAML stdout",
			args:     []string{"cmd", "-to-yaml"},
			input:    string(sshContent),
			expected: yamlContent,
			isPipe:   true,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Cmd.ResetFlags() // Reset flags before each test
			os.Args = tt.args

			if !tt.isPipe {
				err = os.WriteFile(tt.args[2], []byte(tt.input), 0644)
				if err != nil {
					t.Fatal(err)
				}
				defer os.Remove(tt.args[2])
				defer os.Remove(tt.args[4])
			}

			r, w, _ := os.Pipe()
			os.Stdout = w

			if tt.isPipe {
				pipeR, pipeW, _ := os.Pipe()
				os.Stdin = pipeR
				go func() {
					pipeW.Write([]byte(tt.input))
					pipeW.Close()
				}()
			}

			main()

			w.Close()
			out, _ := io.ReadAll(r)

			if tt.wantErr {
				if !strings.Contains(string(out), "Error") {
					t.Errorf("expected error, got: %s", string(out))
				}
			} else {
				if !tt.isPipe {
					if !strings.Contains(string(out), "File has been saved successfully") {
						t.Errorf("unexpected output:\nexpected to contain: File has been saved successfully\ngot: %s", string(out))
					}

					checkFileContent(t, tt.args[4], tt.expected)
				} else {
					if string(Fn.TidyLastEmptyLines(out)) != string(Fn.TidyLastEmptyLines((tt.expected))) {
						t.Errorf("!!!!unexpected output:\nexpected: %s\ngot: %s", tt.expected, string(out))
					}
				}
			}
		})
	}
}

func checkFileContent(t *testing.T, filename string, expected []byte) {
	t.Helper()
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Errorf("Failed to read output file: %v", err)
	}

	if string(content) != string(expected) {
		t.Errorf("unexpected file content:\nexpected: %s\ngot: %s", expected, string(content))
	}
}
