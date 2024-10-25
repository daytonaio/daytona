// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"io"

	log "github.com/sirupsen/logrus"
)

type logFormatter struct {
	textFormatter  *log.TextFormatter
	agentLogWriter io.Writer
}

func (f *logFormatter) Format(entry *log.Entry) ([]byte, error) {
	formatted, err := f.textFormatter.Format(entry)
	if err != nil {
		return nil, err
	}

	if f.agentLogWriter != nil {
		_, err = f.agentLogWriter.Write(formatted)
		if err != nil {
			return nil, err
		}
	}

	// Return the original message without log decoration
	// We don't want decoration to show up in the target creation logs
	return []byte(entry.Message + "\n"), nil
}

func (s *Agent) initLogs() {
	logFormatter := &logFormatter{
		textFormatter: &log.TextFormatter{
			ForceColors: true,
		},
		agentLogWriter: s.LogWriter,
	}

	log.SetFormatter(logFormatter)
}
