// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package config

import (
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"

	log "github.com/sirupsen/logrus"
)

type Config struct {
	ProjectDir  string
	LogFilePath *string `envconfig:"DAYTONA_DAEMON_LOG_FILE_PATH"`
}

var DEFAULT_LOG_FILE_PATH = "/tmp/daytona-daemon.log"

var config *Config

func GetConfig() (*Config, error) {
	if config != nil {
		return config, nil
	}

	config = &Config{}

	err := envconfig.Process("", config)
	if err != nil {
		log.Error(err)
		os.Exit(2)
	}

	var validate = validator.New()
	err = validate.Struct(config)
	if err != nil {
		return nil, err
	}

	config.LogFilePath = GetLogFilePath()

	return config, nil
}

func GetLogFilePath() *string {
	logFilePath, ok := os.LookupEnv("DAYTONA_DAEMON_LOG_FILE_PATH")
	if !ok {
		return &DEFAULT_LOG_FILE_PATH
	}

	logFilePath = strings.Replace(logFilePath, "(HOME)", os.Getenv("HOME"), 1)

	return &logFilePath
}
