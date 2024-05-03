// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package logger

import (
	"io"
)

type Logger interface {
	io.WriteCloser
	Cleanup() error
}

type LoggerFactory interface {
	CreateWorkspaceLogger(workspaceId string) Logger
	CreateProjectLogger(workspaceId, projectName string) Logger
	CreateWorkspaceLogReader(workspaceId string) (io.Reader, error)
	CreateProjectLogReader(workspaceId, projectName string) (io.Reader, error)
}

type loggerFactoryImpl struct {
	logsDir string
}

func NewLoggerFactory(logsDir string) LoggerFactory {
	return &loggerFactoryImpl{logsDir: logsDir}
}
