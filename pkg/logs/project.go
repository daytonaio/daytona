// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package logs

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

type projectLogger struct {
	logsDir     string
	workspaceId string
	projectName string
	logFile     *os.File
	logger      *logrus.Logger
	source      LogSource
}

func (pl *projectLogger) Write(p []byte) (n int, err error) {
	if pl.logFile == nil {
		filePath := filepath.Join(pl.logsDir, pl.workspaceId, pl.projectName, "log")
		err = os.MkdirAll(filepath.Dir(filePath), 0755)
		if err != nil {
			return len(p), err
		}

		logFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return len(p), err
		}
		pl.logFile = logFile
		pl.logger.SetOutput(pl.logFile)
	}

	var entry LogEntry
	entry.Msg = string(p)
	entry.Source = string(pl.source)
	entry.WorkspaceId = &pl.workspaceId
	entry.ProjectName = &pl.projectName
	entry.Time = time.Now().Format(time.RFC3339)

	b, err := json.Marshal(entry)
	if err != nil {
		return len(p), err
	}

	b = append(b, []byte(LogDelimiter)...)

	_, err = pl.logFile.Write(b)
	if err != nil {
		return len(p), err
	}

	return len(p), nil
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

func (l *loggerFactoryImpl) CreateProjectLogger(workspaceId, projectName string, source LogSource) Logger {
	logger := logrus.New()

	return &projectLogger{
		workspaceId: workspaceId,
		logsDir:     l.wsLogsDir,
		projectName: projectName,
		logger:      logger,
		source:      source,
	}
}

func (l *loggerFactoryImpl) CreateProjectLogReader(workspaceId, projectName string) (io.Reader, error) {
	filePath := filepath.Join(l.wsLogsDir, workspaceId, projectName, "log")
	return os.Open(filePath)
}
