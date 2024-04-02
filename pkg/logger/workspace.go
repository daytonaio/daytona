// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package logger

import (
	"io"
	"os"
	"path/filepath"
)

type workspaceLogger struct {
	logsDir     string
	workspaceId string
	logFile     *os.File
}

func (w *workspaceLogger) Write(p []byte) (n int, err error) {
	if w.logFile == nil {
		filePath := filepath.Join(w.logsDir, w.workspaceId, "log")
		err = os.MkdirAll(filepath.Dir(filePath), 0755)
		if err != nil {
			return 0, err
		}
		logFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return 0, err
		}
		w.logFile = logFile
	}

	return w.logFile.Write(p)
}

func (w *workspaceLogger) Close() error {
	if w.logFile != nil {
		err := w.logFile.Close()
		w.logFile = nil
		return err
	}
	return nil
}

func GetWorkspaceLogger(logsDir, workspaceId string) io.WriteCloser {
	return &workspaceLogger{workspaceId: workspaceId, logsDir: logsDir}
}
