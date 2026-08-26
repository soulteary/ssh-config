package integration_test

import (
	"bytes"
	"testing"

	"github.com/soulteary/ssh-config/v3/pkg/sshconfig"
)

func TestV3PublicImportPath(t *testing.T) {
	t.Parallel()

	input := []byte("Host example\n  User deploy\n")
	document, err := sshconfig.Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	output, err := document.MarshalPreserve()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(output, input) {
		t.Fatalf("round trip = %q, want %q", output, input)
	}
}
