// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package logs

import (
	"os"
	"path/filepath"

	"github.com/daytonaio/daytona/pkg/server/config"
	frp_log "github.com/fatedier/frp/pkg/util/log"
	log "github.com/sirupsen/logrus"
)

var logFilePath *string

type logFormatter struct {
	textFormatter *log.TextFormatter
	file          *os.File
}

func (f *logFormatter) Format(entry *log.Entry) ([]byte, error) {
	formatted, err := f.textFormatter.Format(entry)
	if err != nil {
		return nil, err
	}

	if logFilePath != nil {
		// Write to file
		_, err = f.file.Write(formatted)
		if err != nil {
			return nil, err
		}
	}

	return formatted, nil
}

func GetLogFilePath() *string {
	if logFilePath == nil {
		configDir, err := config.GetConfigDir()
		if err != nil {
			return nil
		}

		filePath := filepath.Join(configDir, "daytona.log")
		logFilePath = &filePath
	}

	return logFilePath
}

func Init() error {
	filePath := GetLogFilePath()

	file, err := os.OpenFile(*filePath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
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

	frpOutput := *filePath
	if os.Getenv("FRP_LOG_OUTPUT") != "" {
		frpOutput = os.Getenv("FRP_LOG_OUTPUT")
	}

	frp_log.InitLog(frpOutput, frpLogLevel, 0, false)

	return nil
}
