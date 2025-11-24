// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package util

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

// PrefixedWriter wraps an io.Writer and adds a prefix to each line
type PrefixedWriter struct {
	Prefix string
	Writer io.Writer
}

func (pw *PrefixedWriter) Write(p []byte) (n int, err error) {
	scanner := bufio.NewScanner(bytes.NewReader(p))
	start := 0
	for scanner.Scan() {
		line := scanner.Text()
		// Find the end of the current line in p
		// Scanner strips the line ending, so we need to find it
		lineBytes := []byte(line)
		lineLen := len(lineBytes)
		// Find the end of the line in p, including the line ending
		end := start + lineLen
		// Advance end to include the line ending(s)
		for end < len(p) && (p[end] == '\n' || p[end] == '\r') {
			end++
		}
		_, err := fmt.Fprintf(pw.Writer, "%s%s\n", pw.Prefix, line)
		if err != nil {
			return end, err
		}
		start = end
	}
	if err := scanner.Err(); err != nil {
		return start, err
	}
	return len(p), nil
}
