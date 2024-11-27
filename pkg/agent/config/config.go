// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"

	log "github.com/sirupsen/logrus"
)

type DaytonaServerConfig struct {
	Url    string `envconfig:"DAYTONA_SERVER_URL" validate:"required"`
	ApiKey string `envconfig:"DAYTONA_SERVER_API_KEY" validate:"required"`
	ApiUrl string `envconfig:"DAYTONA_SERVER_API_URL" validate:"required"`
}

type Config struct {
	WorkspaceDir string
	ClientId     string  `envconfig:"DAYTONA_CLIENT_ID" validate:"required"`
	WorkspaceId  string  `envconfig:"DAYTONA_WORKSPACE_ID"`
	TargetId     string  `envconfig:"DAYTONA_TARGET_ID" validate:"required"`
	LogFilePath  *string `envconfig:"DAYTONA_AGENT_LOG_FILE_PATH"`
	Server       DaytonaServerConfig
	Mode         Mode
}

type Mode string

const (
	ModeTarget    Mode = "target"
	ModeWorkspace Mode = "workspace"
)

var config *Config

func GetConfig(mode Mode) (*Config, error) {
	if config != nil {
		if config.Mode != mode {
			return nil, errors.New("config mode does not match requested mode")
		}
		return config, nil
	}

	config = &Config{
		Mode: mode,
	}

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

	if config.Mode == ModeWorkspace {
		if config.WorkspaceId == "" {
			return nil, errors.New("DAYTONA_WORKSPACE_ID is required in workspace mode")
		}
	}

	config.LogFilePath = GetLogFilePath()

	return config, nil
}

func GetLogFilePath() *string {
	logFilePath, ok := os.LookupEnv("DAYTONA_AGENT_LOG_FILE_PATH")
	if !ok {
		return nil
	}

	logFilePath = strings.Replace(logFilePath, "(HOME)", os.Getenv("HOME"), 1)

	return &logFilePath
}
