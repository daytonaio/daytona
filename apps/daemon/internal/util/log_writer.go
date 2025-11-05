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
	for scanner.Scan() {
		line := scanner.Text()
		_, err := fmt.Fprintf(pw.Writer, "%s%s\n", pw.Prefix, line)
		if err != nil {
			return 0, err
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return len(p), nil
}
