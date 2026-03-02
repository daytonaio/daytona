// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"io"
)

type PrefixWriter struct {
	prefix []byte
	writer io.Writer
}

func NewPrefixWriter(prefix []byte, writer io.Writer) *PrefixWriter {
	return &PrefixWriter{prefix: prefix, writer: writer}
}

func (w *PrefixWriter) Write(p []byte) (n int, err error) {
	logLine := append(w.prefix, p...)
	_, err = w.writer.Write(logLine)
	if err != nil {
		return 0, err
	}

	return len(p), nil
}
