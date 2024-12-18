// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package logs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var LogDelimiter = "!-#_^*|\n"

type Logger interface {
	io.WriteCloser
	ConstructJsonLogEntry(p []byte) ([]byte, error)
	Cleanup() error
}

type LogSource string

const (
	LogSourceServer   LogSource = "server"
	LogSourceProvider LogSource = "provider"
	LogSourceBuilder  LogSource = "builder"
	LogSourceRunner   LogSource = "runner"
)

type LogEntry struct {
	Source string `json:"source"`
	Label  string `json:"label"`
	Msg    string `json:"msg"`
	Level  string `json:"level"`
	Time   string `json:"time"`
}

type ILoggerFactory interface {
	CreateLogger(id, label string, source LogSource) (Logger, error)
	CreateLogReader(id string) (io.Reader, error)
	CreateLogWriter(id string) (io.WriteCloser, error)
}

type loggerFactory struct {
	logsDir string
}

func NewLoggerFactory(logsDir string) ILoggerFactory {
	return &loggerFactory{
		logsDir: logsDir,
	}
}

func (l *loggerFactory) CreateLogger(id, label string, source LogSource) (Logger, error) {
	return &logger{
		id:      id,
		logsDir: l.logsDir,
		label:   label,
		source:  source,
	}, nil
}

func (l *loggerFactory) CreateLogReader(id string) (io.Reader, error) {
	filePath := filepath.Join(l.logsDir, id, "log")

	dirPath := filepath.Dir(filePath)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directories for path %s: %v", dirPath, err)
	}

	return os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0644)
}

func (l *loggerFactory) CreateLogWriter(id string) (io.WriteCloser, error) {
	return &logger{
		id:                   id,
		logsDir:              l.logsDir,
		skipEntryConstructor: true,
	}, nil
}
