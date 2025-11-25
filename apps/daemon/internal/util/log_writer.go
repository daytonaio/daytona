// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package util

import (
	"bytes"
	"fmt"
	"io"
)

// PrefixedWriter wraps an io.Writer and adds a prefix to each line
type PrefixedWriter struct {
	Prefix string
	Writer io.Writer
	buf    bytes.Buffer // buffer to store partial lines between writes
}

func (pw *PrefixedWriter) Write(p []byte) (n int, err error) {
	// Prepend any buffered partial line to the new input
	var data []byte
	if pw.buf.Len() > 0 {
		data = append(pw.buf.Bytes(), p...)
		pw.buf.Reset()
	} else {
		data = p
	}
	// Process complete lines
	start := 0
	for i, b := range data {
		if b == '\n' {
			line := data[start:i]
			_, err := fmt.Fprintf(pw.Writer, "%s%s\n", pw.Prefix, line)
			if err != nil {
				return 0, err
			}
			start = i + 1
		}
	}
	// Buffer any remaining partial line
	if start < len(data) {
		pw.buf.Write(data[start:])
	}
	return len(p), nil
}
