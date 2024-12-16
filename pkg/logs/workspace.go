// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package logs

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

type WorkspaceLogger struct {
	logsDir       string
	WorkspaceId   string
	workspaceName string
	logFile       *os.File
	logger        *logrus.Logger
	source        LogSource
}

func (w *WorkspaceLogger) Write(p []byte) (n int, err error) {
	if w.logFile == nil {
		filePath := filepath.Join(w.logsDir, w.WorkspaceId, "log")
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

	b, err := w.ConstructJsonLogEntry(p)
	if err != nil {
		return len(p), err
	}

	_, err = w.logFile.Write(b)
	if err != nil {
		return len(p), err
	}

	return len(p), nil
}

func (w *WorkspaceLogger) ConstructJsonLogEntry(p []byte) ([]byte, error) {
	var entry LogEntry
	entry.Msg = string(p)
	entry.Source = string(w.source)
	entry.WorkspaceName = &w.workspaceName
	entry.Time = time.Now().Format(time.RFC3339)

	b, err := json.Marshal(entry)
	if err != nil {
		return nil, err
	}

	return append(b, []byte(LogDelimiter)...), nil
}

func (w *WorkspaceLogger) Close() error {
	if w.logFile != nil {
		err := w.logFile.Close()
		w.logFile = nil
		return err
	}
	return nil
}

func (w *WorkspaceLogger) Cleanup() error {
	workspaceLogsDir := filepath.Join(w.logsDir, w.WorkspaceId)

	_, err := os.Stat(workspaceLogsDir)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	return os.RemoveAll(workspaceLogsDir)
}

func (l *loggerFactory) CreateWorkspaceLogger(workspaceId, workspaceName string, source LogSource) (Logger, error) {
	logger := logrus.New()

	return &WorkspaceLogger{
		WorkspaceId:   workspaceId,
		logsDir:       l.targetLogsDir,
		workspaceName: workspaceName,
		logger:        logger,
		source:        source,
	}, nil
}

func (l *loggerFactory) CreateWorkspaceLogReader(workspaceId string) (io.Reader, error) {
	filePath := filepath.Join(l.targetLogsDir, workspaceId, "log")

	dirPath := filepath.Dir(filePath)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directories for path %s: %v", dirPath, err)
	}

	return os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0644)
}
