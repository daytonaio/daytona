// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	frp_log "github.com/fatedier/frp/pkg/util/log"
	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

type logFormatter struct {
	textFormatter *log.TextFormatter
	writer        io.Writer
}

func (f *logFormatter) Format(entry *log.Entry) ([]byte, error) {
	formatted, err := f.textFormatter.Format(entry)
	if err != nil {
		return nil, err
	}

	if f.writer != nil {
		// Write to file
		_, err = f.writer.Write(formatted)
		if err != nil {
			return nil, err
		}
	}

	return formatted, nil
}

func (s *Server) initLogs() error {
	rotatedLogFile := &lumberjack.Logger{
		Filename:   s.config.LogFile.Path,
		MaxSize:    s.config.LogFile.MaxSize, // megabytes
		MaxBackups: s.config.LogFile.MaxBackups,
		MaxAge:     s.config.LogFile.MaxAge, // days
		LocalTime:  s.config.LogFile.LocalTime,
		Compress:   s.config.LogFile.Compress,
	}

	logFormatter := &logFormatter{
		textFormatter: &log.TextFormatter{
			ForceColors: true,
		},
		writer: rotatedLogFile,
	}

	log.SetFormatter(logFormatter)

	frpLogLevel := "error"
	if os.Getenv("FRP_LOG_LEVEL") != "" {
		frpLogLevel = os.Getenv("FRP_LOG_LEVEL")
	}

	frpOutput := s.config.LogFile.Path
	if os.Getenv("FRP_LOG_OUTPUT") != "" {
		frpOutput = os.Getenv("FRP_LOG_OUTPUT")
	}

	frp_log.InitLogger(frpOutput, frpLogLevel, 0, false)

	return nil
}

func (s *Server) GetLogReader() (io.Reader, error) {
	file, err := os.Open(s.config.LogFile.Path)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (s *Server) GetLogFiles() ([]string, error) {
	logDir := filepath.Dir(s.config.LogFile.Path)

	files, err := os.ReadDir(logDir)
	if err != nil {
		return nil, err
	}

	var logFiles []string
	for _, file := range files {
		if file.Name() == "daytona.log" || strings.HasPrefix(file.Name(), "daytona-") && (strings.HasSuffix(file.Name(), ".log") || strings.HasSuffix(file.Name(), ".zip") || strings.HasSuffix(file.Name(), ".gz")) {
			logFiles = append(logFiles, filepath.Join(logDir, file.Name()))
		}
	}

	return logFiles, nil
}
