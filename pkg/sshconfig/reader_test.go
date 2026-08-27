package sshconfig

import (
	"io"
	"strings"
	"testing"
)

func TestParseReaderLimitsInput(t *testing.T) {
	t.Parallel()
	input := "Host example\n"

	if _, err := ParseReader(strings.NewReader(input), ParseOptions{MaxBytes: int64(len(input))}); err != nil {
		t.Fatalf("exact limit: %v", err)
	}
	if _, err := ParseReader(strings.NewReader(input), ParseOptions{MaxBytes: int64(len(input) - 1)}); err == nil || !strings.Contains(err.Error(), "exceeds maximum size") {
		t.Fatalf("over limit error = %v", err)
	}
	if _, err := ParseReader(strings.NewReader(input), ParseOptions{}); err != nil {
		t.Fatalf("unlimited input: %v", err)
	}
	if _, err := ParseReader(nil, ParseOptions{}); err == nil || !strings.Contains(err.Error(), "nil") {
		t.Fatalf("nil reader error = %v", err)
	}
	if _, err := ParseReader(strings.NewReader(input), ParseOptions{MaxBytes: -1}); err == nil || !strings.Contains(err.Error(), "negative") {
		t.Fatalf("negative limit error = %v", err)
	}
}

type countingReader struct {
	reader io.Reader
	read   int
}

func (reader *countingReader) Read(buffer []byte) (int, error) {
	count, err := reader.reader.Read(buffer)
	reader.read += count
	return count, err
}

func TestReadAllAtMostStopsAfterTheBudget(t *testing.T) {
	t.Parallel()
	reader := &countingReader{reader: strings.NewReader(strings.Repeat("x", 1024*1024))}
	_, exceeded, err := readAllAtMost(reader, 8)
	if err != nil {
		t.Fatal(err)
	}
	if !exceeded {
		t.Fatal("readAllAtMost() accepted an oversized stream")
	}
	if reader.read != 9 {
		t.Fatalf("readAllAtMost() consumed %d bytes, want the 8-byte budget plus one probe", reader.read)
	}
}
