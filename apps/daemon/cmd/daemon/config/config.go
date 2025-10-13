// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package config

import (
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"

	log "github.com/sirupsen/logrus"
)

type Config struct {
	DaemonLogFilePath            string `envconfig:"DAYTONA_DAEMON_LOG_FILE_PATH"`
	EntrypointLogFilePath        string `envconfig:"DAYTONA_ENTRYPOINT_LOG_FILE_PATH"`
	EntrypointShutdownTimeoutSec int    `envconfig:"ENTRYPOINT_SHUTDOWN_TIMEOUT_SEC"`
	SigtermShutdownTimeoutSec    int    `envconfig:"SIGTERM_SHUTDOWN_TIMEOUT_SEC"`
	UserHomeAsWorkDir            bool   `envconfig:"DAYTONA_USER_HOME_AS_WORKDIR"`
	SandboxId    string  `envconfig:"DAYTONA_SANDBOX_ID" validate:"required"`
	OtelEndpoint string  `envconfig:"DAYTONA_OTEL_ENDPOINT" validate:"required,url"`
}

var defaultDaemonLogFilePath = "/tmp/daytona-daemon.log"
var defaultEntrypointLogFilePath = "/tmp/daytona-entrypoint.log"

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

	if config.DaemonLogFilePath == "" {
		config.DaemonLogFilePath = defaultDaemonLogFilePath
	}

	if config.EntrypointLogFilePath == "" {
		config.EntrypointLogFilePath = defaultEntrypointLogFilePath
	}

	if config.EntrypointShutdownTimeoutSec <= 0 {
		// Default to 10 seconds
		config.EntrypointShutdownTimeoutSec = 10
	}

	if config.SigtermShutdownTimeoutSec <= 0 {
		// Default to 5 seconds
		config.SigtermShutdownTimeoutSec = 5
	}

	return config, nil
}
