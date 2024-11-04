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

type workspaceLogger struct {
	logsDir       string
	workspaceId   string
	workspaceName string
	logFile       *os.File
	logger        *logrus.Logger
	source        LogSource
}

func (pl *workspaceLogger) Write(p []byte) (n int, err error) {
	if pl.logFile == nil {
		filePath := filepath.Join(pl.logsDir, pl.workspaceId, "log")
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
	entry.WorkspaceName = &pl.workspaceName
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

func (pl *workspaceLogger) Close() error {
	if pl.logFile != nil {
		err := pl.logFile.Close()
		pl.logFile = nil
		return err
	}
	return nil
}

func (pl *workspaceLogger) Cleanup() error {
	workspaceLogsDir := filepath.Join(pl.logsDir, pl.workspaceId)

	_, err := os.Stat(workspaceLogsDir)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	return os.RemoveAll(workspaceLogsDir)
}

func (l *loggerFactoryImpl) CreateWorkspaceLogger(workspaceId, workspaceName string, source LogSource) Logger {
	logger := logrus.New()

	return &workspaceLogger{
		workspaceId:   workspaceId,
		logsDir:       l.targetLogsDir,
		workspaceName: workspaceName,
		logger:        logger,
		source:        source,
	}
}

func (l *loggerFactoryImpl) CreateWorkspaceLogReader(workspaceId string) (io.Reader, error) {
	filePath := filepath.Join(l.targetLogsDir, workspaceId, "log")
	return os.Open(filePath)
}
