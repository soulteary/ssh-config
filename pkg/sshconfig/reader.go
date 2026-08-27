package sshconfig

import (
	"fmt"
	"io"
)

// ParseOptions controls resource use while reading a lossless document.
type ParseOptions struct {
	// MaxBytes rejects larger inputs. Zero means unlimited.
	MaxBytes int64
}

// ParseReader reads and parses a lossless document. Use MaxBytes when the
// source is not trusted by the application.
func ParseReader(reader io.Reader, options ParseOptions) (*Document, error) {
	if reader == nil {
		return nil, fmt.Errorf("sshconfig: input reader is nil")
	}
	if options.MaxBytes < 0 {
		return nil, fmt.Errorf("sshconfig: MaxBytes must not be negative")
	}
	input, err := readAllLimited(reader, options.MaxBytes, "input")
	if err != nil {
		return nil, err
	}
	return Parse(input)
}

func readAllLimited(reader io.Reader, maxBytes int64, subject string) ([]byte, error) {
	if maxBytes <= 0 {
		return io.ReadAll(reader)
	}

	limited := &io.LimitedReader{R: reader, N: maxBytes}
	input, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(input)) < maxBytes {
		return input, nil
	}

	var probe [1]byte
	n, probeErr := io.ReadFull(reader, probe[:])
	if n > 0 {
		return nil, fmt.Errorf("sshconfig: %s exceeds maximum size of %d bytes", subject, maxBytes)
	}
	if probeErr != nil && probeErr != io.EOF {
		return nil, probeErr
	}
	return input, nil
}
