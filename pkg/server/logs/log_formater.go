// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package logs

import (
	"os"
	"path"

	"github.com/daytonaio/daytona/pkg/server/config"
	log "github.com/sirupsen/logrus"
)

var LogFilePath *string

type logFormatter struct {
	textFormater *log.TextFormatter
	file         *os.File
}

func (f *logFormatter) Format(entry *log.Entry) ([]byte, error) {
	formatted, err := f.textFormater.Format(entry)
	if err != nil {
		return nil, err
	}

	if LogFilePath != nil {
		// Write to file
		_, err = f.file.Write(formatted)
		if err != nil {
			return nil, err
		}
	}

	return formatted, nil
}

func Init() error {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return err
	}

	filePath := path.Join(configDir, "daytona.log")
	LogFilePath = &filePath

	file, err := os.OpenFile(*LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	logFormatter := &logFormatter{
		textFormater: new(log.TextFormatter),
		file:         file,
	}

	log.SetFormatter(logFormatter)

	return nil
}
