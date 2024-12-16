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

type targetLogger struct {
	logsDir    string
	targetId   string
	targetName string
	logFile    *os.File
	logger     *logrus.Logger
	source     LogSource
}

func (t *targetLogger) Write(p []byte) (n int, err error) {
	if t.logFile == nil {
		filePath := filepath.Join(t.logsDir, t.targetId, "log")
		err = os.MkdirAll(filepath.Dir(filePath), 0755)
		if err != nil {
			return len(p), err
		}
		logFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return len(p), err
		}
		t.logFile = logFile
		t.logger.SetOutput(t.logFile)
	}

	b, err := t.ConstructJsonLogEntry(p)
	if err != nil {
		return len(p), err
	}

	_, err = t.logFile.Write(b)
	if err != nil {
		return len(p), err
	}

	return len(p), nil
}

func (t *targetLogger) ConstructJsonLogEntry(p []byte) ([]byte, error) {
	var entry LogEntry
	entry.Msg = string(p)
	entry.Source = string(t.source)
	entry.TargetName = &t.targetName
	entry.Time = time.Now().Format(time.RFC3339)

	b, err := json.Marshal(entry)
	if err != nil {
		return nil, err
	}

	return append(b, []byte(LogDelimiter)...), nil
}

func (t *targetLogger) Close() error {
	if t.logFile != nil {
		err := t.logFile.Close()
		t.logFile = nil
		return err
	}
	return nil
}

func (t *targetLogger) Cleanup() error {
	targetLogsDir := filepath.Join(t.logsDir, t.targetId)

	_, err := os.Stat(targetLogsDir)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	return os.RemoveAll(targetLogsDir)
}

func (l *loggerFactory) CreateTargetLogger(targetId, targetName string, source LogSource) (Logger, error) {
	logger := logrus.New()

	return &targetLogger{
		targetId:   targetId,
		targetName: targetName,
		logsDir:    l.targetLogsDir,
		logger:     logger,
		source:     source,
	}, nil
}

func (l *loggerFactory) CreateTargetLogReader(targetId string) (io.Reader, error) {
	filePath := filepath.Join(l.targetLogsDir, targetId, "log")
	return os.Open(filePath)
}
