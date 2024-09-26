// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"io"
	"os"

	frp_log "github.com/fatedier/frp/pkg/util/log"
	log "github.com/sirupsen/logrus"
)

type logFormatter struct {
	textFormatter *log.TextFormatter
	file          *os.File
}

func (f *logFormatter) Format(entry *log.Entry) ([]byte, error) {
	formatted, err := f.textFormatter.Format(entry)
	if err != nil {
		return nil, err
	}

	if f.file != nil {
		// Write to file
		_, err = f.file.Write(formatted)
		if err != nil {
			return nil, err
		}
	}

	return formatted, nil
}

func (s *Server) initLogs() error {
	filePath := s.config.LogFilePath

	file, err := os.OpenFile(filePath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	logFormatter := &logFormatter{
		textFormatter: &log.TextFormatter{
			ForceColors: true,
		},
		file: file,
	}

	log.SetFormatter(logFormatter)

	frpLogLevel := "error"
	if os.Getenv("FRP_LOG_LEVEL") != "" {
		frpLogLevel = os.Getenv("FRP_LOG_LEVEL")
	}

	frpOutput := filePath
	if os.Getenv("FRP_LOG_OUTPUT") != "" {
		frpOutput = os.Getenv("FRP_LOG_OUTPUT")
	}

	frp_log.InitLogger(frpOutput, frpLogLevel, 0, false)

	return nil
}

func (s *Server) GetLogReader() (io.Reader, error) {
	file, err := os.Open(s.config.LogFilePath)
	if err != nil {
		return nil, err
	}

	return file, nil
}
