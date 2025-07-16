// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package terminal

import (
	"bytes"
	"unicode/utf8"
)

type UTF8Decoder struct {
	buffer []byte
}

func NewUTF8Decoder() *UTF8Decoder {
	return &UTF8Decoder{
		buffer: make([]byte, 0, 1024),
	}
}

// Write appends new data to the internal buffer and decodes valid UTF-8 runes.
// It returns the decoded string. Any incomplete bytes are kept for the next call.
func (d *UTF8Decoder) Write(data []byte) string {
	// Combine buffer + new data
	data = append(d.buffer, data...)
	var output bytes.Buffer

	i := 0
	for i < len(data) {
		r, size := utf8.DecodeRune(data[i:])
		if r == utf8.RuneError {
			if size == 1 {
				// Could be incomplete rune at the end
				remaining := len(data) - i
				if remaining < utf8.UTFMax {
					// Buffer the remaining bytes for next call
					break
				}
				// Otherwise, it's an invalid byte, emit replacement and advance by 1
				output.WriteRune(r)
				i++
				continue
			}
		}
		output.WriteRune(r)
		i += size
	}

	// Save leftover bytes (possibly an incomplete rune)
	d.buffer = d.buffer[:0]
	if i < len(data) {
		d.buffer = append(d.buffer, data[i:]...)
	}

	return output.String()
}
