// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package util

import (
	"io"

	log "github.com/sirupsen/logrus"
)

type LogFormatter struct {
	TextFormatter    *log.TextFormatter
	ProcessLogWriter io.Writer
}

func (f *LogFormatter) Format(entry *log.Entry) ([]byte, error) {
	formatted, err := f.TextFormatter.Format(entry)
	if err != nil {
		return nil, err
	}

	if f.ProcessLogWriter != nil {
		_, err = f.ProcessLogWriter.Write(formatted)
		if err != nil {
			return nil, err
		}
	}

	// Return the original message without log decoration
	// We don't want decoration to show up in the target creation logs
	return []byte(entry.Message + "\n"), nil
}
