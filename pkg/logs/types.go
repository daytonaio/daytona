// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package logs

import (
	"fmt"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
)

var defaultLogFileConfig = LogFileConfig{
	MaxSize:    100, // megabytes
	MaxBackups: 7,
	MaxAge:     15, // days
	LocalTime:  true,
	Compress:   true,
}

type LogFileConfig struct {
	Path       string `json:"path" validate:"required"`
	MaxSize    int    `json:"maxSize" validate:"required"`
	MaxBackups int    `json:"maxBackups" validate:"required"`
	MaxAge     int    `json:"maxAge" validate:"required"`
	LocalTime  bool   `json:"localTime" validate:"optional"`
	Compress   bool   `json:"compress" validate:"optional"`
} // @name LogFileConfig

func GetDefaultLogFileConfig(logFilePath string) *LogFileConfig {
	logFileConfig := LogFileConfig{
		Path:       logFilePath,
		MaxSize:    defaultLogFileConfig.MaxSize,
		MaxBackups: defaultLogFileConfig.MaxBackups,
		MaxAge:     defaultLogFileConfig.MaxAge,
		LocalTime:  defaultLogFileConfig.LocalTime,
		Compress:   defaultLogFileConfig.Compress,
	}

	logFileMaxSize := os.Getenv("DEFAULT_LOG_FILE_MAX_SIZE")
	if logFileMaxSize != "" {
		value, err := strconv.Atoi(logFileMaxSize)
		if err != nil {
			log.Error(fmt.Printf("%s. Using default log file max size.", err))
		} else {
			logFileConfig.MaxSize = value
		}
	}

	logFileMaxBackups := os.Getenv("DEFAULT_LOG_FILE_MAX_BACKUPS")
	if logFileMaxBackups != "" {
		value, err := strconv.Atoi(logFileMaxBackups)
		if err != nil {
			log.Error(fmt.Printf("%s. Using default log file max backups.", err))
		} else {
			logFileConfig.MaxBackups = value
		}
	}

	logFileMaxAge := os.Getenv("DEFAULT_LOG_FILE_MAX_AGE")
	if logFileMaxAge != "" {
		value, err := strconv.Atoi(logFileMaxAge)
		if err != nil {
			log.Error(fmt.Printf("%s. Using default log file max age.", err))
		} else {
			logFileConfig.MaxAge = value
		}
	}

	logFileLocalTime := os.Getenv("DEFAULT_LOG_FILE_LOCAL_TIME")
	if logFileLocalTime != "" {
		value, err := strconv.ParseBool(logFileLocalTime)
		if err != nil {
			log.Error(fmt.Printf("%s. Using default log file local time.", err))
		} else {
			logFileConfig.LocalTime = value
		}
	}

	logFileCompress := os.Getenv("DEFAULT_LOG_FILE_COMPRESS")
	if logFileCompress != "" {
		value, err := strconv.ParseBool(logFileCompress)
		if err != nil {
			log.Error(fmt.Printf("%s. Using default log file compress.", err))
		} else {
			logFileConfig.Compress = value
		}
	}

	return &logFileConfig
}
