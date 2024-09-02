// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package logs

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

type buildLogger struct {
	logsDir string
	buildId string
	logFile *os.File
	logger  *logrus.Logger
	source  LogSource
}

func (bl *buildLogger) Write(p []byte) (n int, err error) {
	if bl.logFile == nil {
		filePath := filepath.Join(bl.logsDir, bl.buildId, "log")
		err = os.MkdirAll(filepath.Dir(filePath), 0755)
		if err != nil {
			return len(p), err
		}

		logFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return len(p), err
		}
		bl.logFile = logFile
		bl.logger.SetOutput(bl.logFile)
	}

	var entry LogEntry
	entry.Msg = string(p)
	entry.Source = string(bl.source)
	entry.BuildId = &bl.buildId

	b, err := json.Marshal(entry)
	if err != nil {
		return len(p), err
	}

	b = append(b, []byte(LogDelimiter)...)

	_, err = bl.logFile.Write(b)
	if err != nil {
		return len(p), err
	}

	return len(p), nil
}

func (bl *buildLogger) Close() error {
	if bl.logFile != nil {
		err := bl.logFile.Close()
		bl.logFile = nil
		return err
	}
	return nil
}

func (bl *buildLogger) Cleanup() error {
	buildLogsDir := filepath.Join(bl.logsDir, bl.buildId)

	_, err := os.Stat(buildLogsDir)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	return os.RemoveAll(buildLogsDir)
}

func (l *loggerFactoryImpl) CreateBuildLogger(buildId string, source LogSource) Logger {
	logger := logrus.New()

	return &buildLogger{
		logsDir: l.buildLogsDir,
		buildId: buildId,
		logger:  logger,
		source:  source,
	}
}

func (l *loggerFactoryImpl) CreateBuildLogReader(buildId string) (io.Reader, error) {
	filePath := filepath.Join(l.buildLogsDir, buildId, "log")
	return os.Open(filePath)
}
