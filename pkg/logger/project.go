// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package logger

import (
	"os"
	"path/filepath"
)

type projectLogger struct {
	logsDir     string
	workspaceId string
	projectName string
	logFile     *os.File
}

func (pl *projectLogger) Write(p []byte) (n int, err error) {
	if pl.logFile == nil {
		filePath := filepath.Join(pl.logsDir, pl.workspaceId, pl.projectName, "log")
		err = os.MkdirAll(filepath.Dir(filePath), 0755)
		if err != nil {
			return 0, err
		}

		logFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return 0, err
		}
		pl.logFile = logFile
	}

	return pl.logFile.Write(p)
}

func (pl *projectLogger) Close() error {
	if pl.logFile != nil {
		err := pl.logFile.Close()
		pl.logFile = nil
		return err
	}
	return nil
}

func (pl *projectLogger) Cleanup() error {
	projectLogsDir := filepath.Join(pl.logsDir, pl.workspaceId, pl.projectName)

	_, err := os.Stat(projectLogsDir)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	return os.RemoveAll(projectLogsDir)
}

func NewProjectLogger(logsDir, workspaceId, projectName string) Logger {
	return &projectLogger{workspaceId: workspaceId, logsDir: logsDir, projectName: projectName}
}
