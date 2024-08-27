//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"io"
	"strings"
)

type PipeReader struct {
	io.ReadCloser
	ExecStream [][]byte
	Index      int
}

func (er *PipeReader) Read(p []byte) (n int, err error) {
	if er.Index >= len(er.ExecStream) {
		return 0, io.EOF
	}
	n = copy(p, er.ExecStream[er.Index])
	er.Index++
	return n, nil
}

func (er *PipeReader) Close() error {
	return nil
}

func NewPipeReader(result string) *PipeReader {
	split := strings.Split(result, "\n")
	execStream := make([][]byte, len(split))
	for i, s := range split {
		execStream[i] = []byte(s)
	}

	return &PipeReader{
		ExecStream: execStream,
		Index:      0,
	}
}
