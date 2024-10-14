package cmd_test

import (
	"bytes"
	"errors"
	"flag"
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

func TestParseArgs(t *testing.T) {
	// 保存原始的 os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name     string
		args     []string
		expected Cmd.Args
	}{
		{
			name: "Test to-yaml flag",
			args: []string{"cmd", "-to-yaml"},
			expected: Cmd.Args{
				ToYAML: true,
			},
		},
		{
			name: "Test to-ssh flag",
			args: []string{"cmd", "-to-ssh"},
			expected: Cmd.Args{
				ToSSH: true,
			},
		},
		{
			name: "Test to-json flag",
			args: []string{"cmd", "-to-json"},
			expected: Cmd.Args{
				ToJSON: true,
			},
		},
		{
			name: "Test src and dest flags",
			args: []string{"cmd", "-src", "input.txt", "-dest", "output.txt"},
			expected: Cmd.Args{
				Src:  "input.txt",
				Dest: "output.txt",
			},
		},
		{
			name: "Test help flag",
			args: []string{"cmd", "-help"},
			expected: Cmd.Args{
				ShowHelp: true,
			},
		},
		{
			name: "Test multiple flags",
			args: []string{"cmd", "-to-yaml", "-src", "input.yaml", "-dest", "output.txt"},
			expected: Cmd.Args{
				ToYAML: true,
				Src:    "input.yaml",
				Dest:   "output.txt",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置命令行参数
			os.Args = tt.args

			// 重置 flag 包的状态，因为 flag.Parse() 只能被调用一次
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			result := Cmd.ParseArgs()

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseArgs() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCheckConvertArgvValid(t *testing.T) {
	tests := []struct {
		name     string
		args     Cmd.Args
		wantBool bool
		wantDesc string
	}{
		{
			name:     "All false",
			args:     Cmd.Args{ToJSON: false, ToSSH: false, ToYAML: false},
			wantBool: true,
			wantDesc: "",
		},
		{
			name:     "Only ToJSON true",
			args:     Cmd.Args{ToJSON: true, ToSSH: false, ToYAML: false},
			wantBool: true,
			wantDesc: "",
		},
		{
			name:     "Only ToSSH true",
			args:     Cmd.Args{ToJSON: false, ToSSH: true, ToYAML: false},
			wantBool: true,
			wantDesc: "",
		},
		{
			name:     "Only ToYAML true",
			args:     Cmd.Args{ToJSON: false, ToSSH: false, ToYAML: true},
			wantBool: true,
			wantDesc: "",
		},
		{
			name:     "ToJSON and ToSSH true",
			args:     Cmd.Args{ToJSON: true, ToSSH: true, ToYAML: false},
			wantBool: false,
			wantDesc: "Please specify either -to-yaml or -to-ssh or -to-json",
		},
		{
			name:     "ToJSON and ToYAML true",
			args:     Cmd.Args{ToJSON: true, ToSSH: false, ToYAML: true},
			wantBool: false,
			wantDesc: "Please specify either -to-yaml or -to-ssh or -to-json",
		},
		{
			name:     "ToSSH and ToYAML true",
			args:     Cmd.Args{ToJSON: false, ToSSH: true, ToYAML: true},
			wantBool: false,
			wantDesc: "Please specify either -to-yaml or -to-ssh or -to-json",
		},
		{
			name:     "All true",
			args:     Cmd.Args{ToJSON: true, ToSSH: true, ToYAML: true},
			wantBool: false,
			wantDesc: "Please specify either -to-yaml or -to-ssh or -to-json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBool, gotDesc := Cmd.CheckConvertArgvValid(tt.args)
			if gotBool != tt.wantBool {
				t.Errorf("CheckConvertArgvValid() gotBool = %v, want %v", gotBool, tt.wantBool)
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
