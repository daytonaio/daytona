//go:build testing

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package logger

import (
	"io"

	"github.com/daytonaio/daytona/pkg/logger"
)

type PipeLogger struct {
	writer io.WriteCloser
}

func (pl *PipeLogger) Write(p []byte) (n int, err error) {
	return pl.writer.Write(p)
}

func (pl *PipeLogger) Close() error {
	return nil
}

func (pl *PipeLogger) Cleanup() error {
	return pl.writer.Close()
}

func NewPipeLogger(writer io.WriteCloser) logger.Logger {
	return &PipeLogger{writer: writer}
}

func NewPipeLogReader(reader io.Reader) (io.Reader, error) {
	return reader, nil
}
