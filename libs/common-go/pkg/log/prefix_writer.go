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
	// Build a fresh buffer so we don't mutate w.prefix.
	logLine := make([]byte, len(w.prefix)+len(p))
	copy(logLine, w.prefix)
	copy(logLine[len(w.prefix):], p)

	written, writeErr := w.writer.Write(logLine)

	// If fewer than the prefix bytes were written, no payload bytes were written.
	if written < len(w.prefix) {
		if writeErr == nil {
			writeErr = io.ErrShortWrite
		}
		return 0, writeErr
	}

	// Compute how many payload bytes were written.
	payloadWritten := written - len(w.prefix)
	if payloadWritten > len(p) {
		payloadWritten = len(p)
	}

	// If not all payload bytes were written, treat as short write.
	if payloadWritten < len(p) {
		if writeErr == nil {
			writeErr = io.ErrShortWrite
		}
		return payloadWritten, writeErr
	}

	// All payload bytes written.
	return len(p), writeErr
}
