// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package config

import (
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"

	log "github.com/sirupsen/logrus"
)

type Config struct {
	DaemonLogFilePath         string        `envconfig:"DAYTONA_DAEMON_LOG_FILE_PATH"`
	UserHomeAsWorkDir         bool          `envconfig:"DAYTONA_USER_HOME_AS_WORKDIR"`
	EntrypointShutdownTimeout time.Duration `envconfig:"ENTRYPOINT_SHUTDOWN_TIMEOUT" default:"10s" validate:"min=1s"`
	SigtermShutdownTimeout    time.Duration `envconfig:"SIGTERM_SHUTDOWN_TIMEOUT" default:"5s" validate:"min=1s"`
	TerminationCheckInterval  time.Duration `envconfig:"DAYTONA_TERMINATION_CHECK_INTERVAL" default:"100ms" validate:"min=1ms"`
	TerminationGracePeriod    time.Duration `envconfig:"DAYTONA_TERMINATION_GRACE_PERIOD" default:"5s" validate:"min=1s"`
}

var defaultDaemonLogFilePath = "/tmp/daytona-daemon.log"

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

	return config, nil
}
