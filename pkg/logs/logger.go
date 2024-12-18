// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package logs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type logger struct {
	logsDir              string
	id                   string
	label                string
	logFile              *os.File
	source               LogSource
	skipEntryConstructor bool
}

func (w *logger) Write(p []byte) (n int, err error) {
	if w.logFile == nil {
		filePath := filepath.Join(w.logsDir, w.id, "log")
		err = os.MkdirAll(filepath.Dir(filePath), 0755)
		if err != nil {
			return len(p), err
		}

		logFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return len(p), err
		}
		w.logFile = logFile
	}

	b := p

	if !w.skipEntryConstructor {
		b, err = w.ConstructJsonLogEntry(p)
		if err != nil {
			return len(p), err
		}
	}

	_, err = w.logFile.Write(b)
	if err != nil {
		return len(p), err
	}

	return len(p), nil
}

func (w *logger) ConstructJsonLogEntry(p []byte) ([]byte, error) {
	var entry LogEntry
	entry.Msg = string(p)
	entry.Source = string(w.source)
	entry.Label = w.label
	entry.Time = time.Now().Format(time.RFC3339)

	b, err := json.Marshal(entry)
	if err != nil {
		return nil, err
	}

	return append(b, []byte(LogDelimiter)...), nil
}

func (w *logger) Close() error {
	if w.logFile != nil {
		err := w.logFile.Close()
		w.logFile = nil
		return err
	}
	return nil
}

func (w *logger) Cleanup() error {
	workspaceLogsDir := filepath.Join(w.logsDir, w.id)

	_, err := os.Stat(workspaceLogsDir)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	return os.RemoveAll(workspaceLogsDir)
}
