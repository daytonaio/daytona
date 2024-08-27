// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package logs

import (
	"io"
)

var LogDelimiter = "!-#_^*|\n"

type Logger interface {
	io.WriteCloser
	Cleanup() error
}

type LogSource string

const (
	LogSourceServer   LogSource = "server"
	LogSourceProvider LogSource = "provider"
	LogSourceBuilder  LogSource = "builder"
)

type LogEntry struct {
	Source      string `json:"source"`
	WorkspaceId string `json:"workspaceId"`
	ProjectName string `json:"projectName"`
	Msg         string `json:"msg"`
	Level       string `json:"level"`
	Time        string `json:"time"`
}

type LoggerFactory interface {
	CreateWorkspaceLogger(workspaceId string, source LogSource) Logger
	CreateProjectLogger(workspaceId, projectName string, source LogSource) Logger
	CreateBuildLogger(projectName, buildId string, source LogSource) Logger
	CreateWorkspaceLogReader(workspaceId string) (io.Reader, error)
	CreateProjectLogReader(workspaceId, projectName string) (io.Reader, error)
	CreateBuildLogReader(projectName, buildId string) (io.Reader, error)
}

type loggerFactoryImpl struct {
	logsDir string
}

func NewLoggerFactory(logsDir string) LoggerFactory {
	return &loggerFactoryImpl{logsDir: logsDir}
}
