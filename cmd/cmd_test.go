package cmd_test

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	Cmd "github.com/soulteary/ssh-config/cmd"
)

func TestCheckConvertArgvValid(t *testing.T) {
	tests := []struct {
		name       string
		args       Cmd.Args
		wantResult bool
		wantDesc   string
	}{
		{
			name:       "Only ToJSON is true",
			args:       Cmd.Args{ToJSON: true, ToSSH: false, ToYAML: false},
			wantResult: true,
			wantDesc:   "",
		},
		{
			name:       "Only ToSSH is true",
			args:       Cmd.Args{ToJSON: false, ToSSH: true, ToYAML: false},
			wantResult: true,
			wantDesc:   "",
		},
		{
			name:       "Only ToYAML is true",
			args:       Cmd.Args{ToJSON: false, ToSSH: false, ToYAML: true},
			wantResult: true,
			wantDesc:   "",
		},
		{
			name:       "All flags are false",
			args:       Cmd.Args{ToJSON: false, ToSSH: false, ToYAML: false},
			wantResult: false,
			wantDesc:   "Please specify either -to-yaml or -to-ssh or -to-json",
		},
		{
			name:       "Multiple flags are true",
			args:       Cmd.Args{ToJSON: true, ToSSH: true, ToYAML: false},
			wantResult: false,
			wantDesc:   "Please specify either -to-yaml or -to-ssh or -to-json",
		},
		{
			name:       "All flags are true",
			args:       Cmd.Args{ToJSON: true, ToSSH: true, ToYAML: true},
			wantResult: false,
			wantDesc:   "Please specify either -to-yaml or -to-ssh or -to-json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, gotDesc := Cmd.CheckConvertArgvValid(tt.args)
			if gotResult != tt.wantResult {
				t.Errorf("CheckConvertArgvValid() gotResult = %v, want %v", gotResult, tt.wantResult)
			}
			if gotDesc != tt.wantDesc {
				t.Errorf("CheckConvertArgvValid() gotDesc = %v, want %v", gotDesc, tt.wantDesc)
			}
		})
	}
}

func TestCheckIOArgvValid(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "testfile.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}
	testDir := filepath.Join(tempDir, "testdir")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		args       Cmd.Args
		wantResult bool
		wantDesc   string
	}{
		{
			name:       "Valid file to file",
			args:       Cmd.Args{Src: testFile, Dest: filepath.Join(tempDir, "newfile.txt")},
			wantResult: true,
			wantDesc:   "",
		},
		{
			name:       "Valid file to existing directory",
			args:       Cmd.Args{Src: testFile, Dest: testDir},
			wantResult: true,
			wantDesc:   "",
		},
		{
			name:       "Valid directory to non-existent directory",
			args:       Cmd.Args{Src: testDir, Dest: filepath.Join(tempDir, "newdir")},
			wantResult: true,
			wantDesc:   "",
		},
		{
			name:       "Invalid: Empty source",
			args:       Cmd.Args{Src: "", Dest: testFile},
			wantResult: false,
			wantDesc:   "Please specify source and destination file path",
		},
		{
			name:       "Invalid: Non-existent source",
			args:       Cmd.Args{Src: filepath.Join(tempDir, "nonexistent"), Dest: testFile},
			wantResult: false,
			wantDesc:   "Error: Source path '" + filepath.Join(tempDir, "nonexistent") + "' does not exist",
		},
		{
			name:       "Invalid: Non-existent parent directory of destination",
			args:       Cmd.Args{Src: testFile, Dest: filepath.Join(tempDir, "nonexistent", "file.txt")},
			wantResult: false,
			wantDesc:   "Error: Parent directory of destination '" + filepath.Join(tempDir, "nonexistent", "file.txt") + "' does not exist",
		},
		{
			name:       "Valid src directory",
			args:       Cmd.Args{Src: testFile},
			wantResult: true,
			wantDesc:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, gotDesc := Cmd.CheckIOArgvValid(tt.args)
			if gotResult != tt.wantResult {
				t.Errorf("CheckIOArgvValid() gotResult = %v, want %v", gotResult, tt.wantResult)
			}
			if gotDesc != tt.wantDesc {
				t.Errorf("CheckIOArgvValid() gotDesc = %v, want %v", gotDesc, tt.wantDesc)
			}
		})
	}
}

type mockFileInfo struct {
	mode os.FileMode
}

func (m mockFileInfo) Name() string       { return "mock" }
func (m mockFileInfo) Size() int64        { return 0 }
func (m mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m mockFileInfo) ModTime() time.Time { return time.Now() }
func (m mockFileInfo) IsDir() bool        { return false }
func (m mockFileInfo) Sys() interface{}   { return nil }

func TestCheckUseStdin(t *testing.T) {
	tests := []struct {
		name     string
		statFunc func() (fs.FileInfo, error)
		want     bool
	}{
		{
			name: "Pipe input",
			statFunc: func() (fs.FileInfo, error) {
				return mockFileInfo{mode: 0}, nil
			},
			want: true,
		},
		{
			name: "Terminal input",
			statFunc: func() (fs.FileInfo, error) {
				return mockFileInfo{mode: os.ModeCharDevice}, nil
			},
			want: false,
		},
		{
			name: "Error case",
			statFunc: func() (fs.FileInfo, error) {
				return nil, errors.New("mock error")
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Cmd.CheckUseStdin(tt.statFunc)
			if got != tt.want {
				t.Errorf("CheckUseStdin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShowHelp(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Cmd.ShowHelp()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	capturedOutput := buf.String()

	expectedOutput := Cmd.Usage

	if strings.TrimSpace(capturedOutput) != strings.TrimSpace(expectedOutput) {
		t.Errorf("ShowHelp() output does not match expected output.\nExpected:\n%s\nGot:\n%s", expectedOutput, capturedOutput)
	}
}

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected Cmd.Args
	}{
		{
			name: "Default values",
			args: []string{},
			expected: Cmd.Args{
				ToYAML:   Cmd.DEFAULT_TO_YAML,
				ToSSH:    Cmd.DEFAULT_TO_SSH,
				ToJSON:   Cmd.DEFAULT_TO_JSON,
				Src:      Cmd.DEFAULT_SRC,
				Dest:     Cmd.DEFAULT_DEST,
				ShowHelp: Cmd.DEFAULT_HELP,
			},
		},
		{
			name: "Set all flags",
			args: []string{"-to-yaml", "-to-ssh", "-to-json", "-src", "source.txt", "-dest", "destination.txt", "-help"},
			expected: Cmd.Args{
				ToYAML:   true,
				ToSSH:    true,
				ToJSON:   true,
				Src:      "source.txt",
				Dest:     "destination.txt",
				ShowHelp: true,
			},
		},
		{
			name: "Set some flags",
			args: []string{"-to-yaml", "-src", "input.yaml"},
			expected: Cmd.Args{
				ToYAML:   true,
				ToSSH:    false,
				ToJSON:   false,
				Src:      "input.yaml",
				Dest:     "",
				ShowHelp: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original os.Args
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			// Set up new os.Args for the test
			os.Args = append([]string{"cmd"}, tt.args...)

			// Reset flags before each test
			Cmd.ResetFlags()

			// Call ParseArgs
			result := Cmd.ParseArgs()

			// Check if the result matches the expected
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseArgs() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestResetFlags(t *testing.T) {
	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set some flags
	os.Args = []string{"cmd", "-to-yaml", "-src", "input.yaml"}
	Cmd.ParseArgs()

	// Reset flags
	Cmd.ResetFlags()

	os.Args = []string{"cmd", "-to-yaml", "-src", "input2.yaml"}

	// Parse args again
	result := Cmd.ParseArgs()

	// Check if all flags are reset to default values
	expected := Cmd.Args{
		ToYAML:   true,
		ToSSH:    Cmd.DEFAULT_TO_SSH,
		ToJSON:   Cmd.DEFAULT_TO_JSON,
		Src:      "input2.yaml",
		Dest:     Cmd.DEFAULT_DEST,
		ShowHelp: Cmd.DEFAULT_HELP,
	}

	if result.Dest != expected.Dest {
		t.Errorf("After ResetFlags(), ParseArgs() = %v, want %v", result, expected)
	}
	if result.ShowHelp != expected.ShowHelp {
		t.Errorf("After ResetFlags(), ParseArgs() = %v, want %v", result, expected)
	}
	if result.Src != expected.Src {
		t.Errorf("After ResetFlags(), ParseArgs() = %v, want %v", result, expected)
	}
	if result.ToJSON != expected.ToJSON {
		t.Errorf("After ResetFlags(), ParseArgs() = %v, want %v", result, expected)
	}
	if result.ToSSH != expected.ToSSH {
		t.Errorf("After ResetFlags(), ParseArgs() = %v, want %v", result, expected)
	}
	if result.ToYAML != expected.ToYAML {
		t.Errorf("After ResetFlags(), ParseArgs() = %v, want %v", result, expected)
	}
}
