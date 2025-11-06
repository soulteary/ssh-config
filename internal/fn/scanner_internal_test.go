package fn

import (
	"bufio"
	"os"
	"strings"
	"testing"
)

func TestReadSingleConfigScannerErrorInternal(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_config_internal_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	longLine := strings.Repeat("a", bufio.MaxScanTokenSize*2)
	content := "Host testhost\n" + longLine

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	if result := ReadSingleConfig(tmpfile.Name()); result != nil {
		t.Fatalf("Expected nil result when scanner error occurs, got: %v", result)
	}
}
