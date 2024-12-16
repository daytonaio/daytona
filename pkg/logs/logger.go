// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package logs

import (
	"io"
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
)

type LogEntry struct {
	Source        string  `json:"source"`
	TargetName    *string `json:"targetName,omitempty"`
	WorkspaceName *string `json:"workspaceName,omitempty"`
	BuildId       *string `json:"buildId,omitempty"`
	Msg           string  `json:"msg"`
	Level         string  `json:"level"`
	Time          string  `json:"time"`
}

type ILoggerFactory interface {
	CreateBuildLogger(buildId string, source LogSource) (Logger, error)
	CreateBuildLogReader(buildId string) (io.Reader, error)

	CreateTargetLogger(targetId, targetName string, source LogSource) (Logger, error)
	CreateTargetLogReader(targetId string) (io.Reader, error)

	CreateWorkspaceLogger(workspaceId, workspaceName string, source LogSource) (Logger, error)
	CreateWorkspaceLogReader(workspaceId string) (io.Reader, error)
}

type loggerFactory struct {
	targetLogsDir string
	buildLogsDir  string
}

func NewLoggerFactory(targetLogsDir *string, buildLogsDir *string) ILoggerFactory {
	loggerFactoryImpl := &loggerFactory{}

	if targetLogsDir != nil {
		loggerFactoryImpl.targetLogsDir = *targetLogsDir
	}

	if buildLogsDir != nil {
		loggerFactoryImpl.buildLogsDir = *buildLogsDir
	}

	return loggerFactoryImpl
}
