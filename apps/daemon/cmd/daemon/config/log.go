// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package config

import (
	"io"

	log "github.com/sirupsen/logrus"
)

type LogFormatter struct {
	TextFormatter *log.TextFormatter
	LogFileWriter io.Writer
}

func (f *LogFormatter) Format(entry *log.Entry) ([]byte, error) {
	formatted, err := f.TextFormatter.Format(entry)
	if err != nil {
		return nil, err
	}

	if f.LogFileWriter != nil {
		_, err = f.LogFileWriter.Write(formatted)
		if err != nil {
			return nil, err
		}
	}

	return []byte(formatted), nil
}
