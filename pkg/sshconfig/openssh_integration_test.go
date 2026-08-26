package sshconfig

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestOpenSSHRoundTripEffectiveConfiguration(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("OpenSSH integration fixture uses POSIX file permissions")
	}
	ssh, err := exec.LookPath("ssh")
	if err != nil {
		t.Skip("OpenSSH client is not installed")
	}
	t.Parallel()

	directory := t.TempDir()
	included := filepath.Join(directory, "included.conf")
	writeTestConfig(t, included, "Host included\n    HostName included.example\n")
	original := filepath.Join(directory, "original.conf")
	generated := filepath.Join(directory, "generated.conf")
	content := []byte("# preserve this comment\n" +
		"Include " + included + "\n" +
		"Host production\n" +
		"    HostName=production.example\n" +
		"    User deploy # inline comment\n" +
		"    IdentityFile ~/.ssh/work\n" +
		"    IdentityFile ~/.ssh/backup\n" +
		"    SetEnv FOO=bar\n" +
		"Host *\n" +
		"    User fallback\n" +
		"    ServerAliveInterval 30\n" +
		"Match host=internal\n" +
		"    HostName internal.example\n")
	writeTestConfig(t, original, string(content))

	doc, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	output, err := doc.MarshalPreserve()
	if err != nil {
		t.Fatalf("MarshalPreserve() error = %v", err)
	}
	writeTestConfig(t, generated, string(output))

	for _, host := range []string{"production", "internal", "included", "other"} {
		host := host
		t.Run(host, func(t *testing.T) {
			originalOutput := runSSHConfig(t, ssh, directory, original, host)
			generatedOutput := runSSHConfig(t, ssh, directory, generated, host)
			if !bytes.Equal(originalOutput, generatedOutput) {
				t.Fatalf("effective configuration changed for %s\n--- original ---\n%s\n--- generated ---\n%s", host, originalOutput, generatedOutput)
			}
		})
	}
}

func TestOpenSSHAcceptsSupportedLexicalForms(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("OpenSSH integration fixture uses POSIX file permissions")
	}
	ssh, err := exec.LookPath("ssh")
	if err != nil {
		t.Skip("OpenSSH client is not installed")
	}
	t.Parallel()

	directory := t.TempDir()
	path := filepath.Join(directory, "config")
	input := []byte("Host=example alias\n" +
		"    HostName = example.test\n" +
		"    ProxyCommand \"ssh -W %h:%p jump\" # comment\n" +
		"    SetEnv FOO=bar BAZ=qux\n")
	doc, err := Parse(input)
	if err != nil || len(doc.Diagnostics()) != 0 {
		t.Fatalf("Parse() = %v, diagnostics = %#v", err, doc.Diagnostics())
	}
	output, _ := doc.MarshalPreserve()
	writeTestConfig(t, path, string(output))
	got := runSSHConfig(t, ssh, directory, path, "example")
	if !bytes.Contains(got, []byte("hostname example.test\n")) {
		t.Fatalf("ssh -G output does not contain parsed hostname:\n%s", got)
	}
}

func TestOpenSSHArgvSemanticsSurviveCanonicalRewrite(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("OpenSSH integration fixture uses POSIX file permissions")
	}
	ssh, err := exec.LookPath("ssh")
	if err != nil {
		t.Skip("OpenSSH client is not installed")
	}
	t.Parallel()

	directory := t.TempDir()
	original := filepath.Join(directory, "original.conf")
	generated := filepath.Join(directory, "generated.conf")
	content := []byte("Host example\n" +
		`    SetEnv FIRST=one" two" SECOND='three'four THIRD=hash#tag FOUR=quote\"mark` + "\n")
	writeTestConfig(t, original, string(content))

	doc, err := Parse(content)
	if err != nil || len(doc.Diagnostics()) != 0 {
		t.Fatalf("Parse() = %v, diagnostics = %#v", err, doc.Diagnostics())
	}
	node := doc.Nodes()[1]
	arguments := make([]string, 0, len(node.Directive.Arguments))
	for _, argument := range node.Directive.Arguments {
		arguments = append(arguments, argument.Value)
	}
	if err := doc.ReplaceDirective(node.ID, "SetEnv", arguments...); err != nil {
		t.Fatal(err)
	}
	output, err := doc.MarshalPreserve()
	if err != nil {
		t.Fatal(err)
	}
	writeTestConfig(t, generated, string(output))

	originalOutput := runSSHConfig(t, ssh, directory, original, "example")
	generatedOutput := runSSHConfig(t, ssh, directory, generated, "example")
	if !bytes.Equal(originalOutput, generatedOutput) {
		t.Fatalf("effective configuration changed after canonical rewrite\n--- original ---\n%s\n--- generated ---\n%s\n--- rewritten source ---\n%s", originalOutput, generatedOutput, output)
	}
}

func runSSHConfig(t *testing.T, ssh, home, config, host string) []byte {
	t.Helper()
	command := exec.Command(ssh, "-G", "-F", config, host)
	command.Env = append(os.Environ(), "HOME="+home)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("ssh -G failed: %v: %s", err, strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes()
}

func FuzzEditMarshalReparse(f *testing.F) {
	f.Add([]byte("Host example\n    User root\n"), "IdentityFile", "~/.ssh/work")
	f.Add([]byte("# comment without newline"), "SetEnv", "FOO=bar")
	f.Fuzz(func(t *testing.T, input []byte, keyword, value string) {
		if keyword == "" || strings.ContainsAny(keyword, " \t\r\n=\"'") {
			t.Skip()
		}
		doc, err := Parse(input)
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		if _, err := doc.AppendDirective(keyword, value); err != nil {
			t.Fatalf("AppendDirective() error = %v", err)
		}
		first, err := doc.MarshalPreserve()
		if err != nil {
			t.Fatalf("MarshalPreserve() error = %v", err)
		}
		reparsed, err := Parse(first)
		if err != nil {
			t.Fatalf("reparse error = %v", err)
		}
		second, err := reparsed.MarshalPreserve()
		if err != nil {
			t.Fatalf("second MarshalPreserve() error = %v", err)
		}
		if !bytes.Equal(first, second) {
			t.Fatalf("serialization is unstable\nfirst: %q\nsecond: %q", first, second)
		}
	})
}
