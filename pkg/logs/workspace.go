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
	logsDir     string
	workspaceId string
	logFile     *os.File
	logger      *logrus.Logger
	source      LogSource
}

func (w *workspaceLogger) Write(p []byte) (n int, err error) {
	if w.logFile == nil {
		filePath := filepath.Join(w.logsDir, w.workspaceId, "log")
		err = os.MkdirAll(filepath.Dir(filePath), 0755)
		if err != nil {
			return len(p), err
		}
		logFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return len(p), err
		}
		w.logFile = logFile
		w.logger.SetOutput(w.logFile)
	}

	var entry LogEntry
	entry.Msg = string(p)
	entry.Source = string(w.source)
	entry.WorkspaceId = &w.workspaceId
	entry.Time = time.Now().Format(time.RFC3339)

	b, err := json.Marshal(entry)
	if err != nil {
		return len(p), err
	}

	b = append(b, []byte(LogDelimiter)...)

	_, err = w.logFile.Write(b)
	if err != nil {
		return len(p), err
	}

	return len(p), nil
}

func (w *workspaceLogger) Close() error {
	if w.logFile != nil {
		err := w.logFile.Close()
		w.logFile = nil
		return err
	}
	return nil
}

func (w *workspaceLogger) Cleanup() error {
	workspaceLogsDir := filepath.Join(w.logsDir, w.workspaceId)

	_, err := os.Stat(workspaceLogsDir)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	return os.RemoveAll(workspaceLogsDir)
}

func (l *loggerFactoryImpl) CreateWorkspaceLogger(workspaceId string, source LogSource) Logger {
	logger := logrus.New()

	return &workspaceLogger{
		workspaceId: workspaceId,
		logsDir:     l.wsLogsDir,
		logger:      logger,
		source:      source,
	}
}

func (l *loggerFactoryImpl) CreateWorkspaceLogReader(workspaceId string) (io.Reader, error) {
	filePath := filepath.Join(l.wsLogsDir, workspaceId, "log")
	return os.Open(filePath)
}
