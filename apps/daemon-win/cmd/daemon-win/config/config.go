// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package config

import (
	"os"
	"path/filepath"

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
}

var config *Config

// getDefaultLogPath returns the default log path for Windows
func getDefaultLogPath(filename string) string {
	// Try TEMP environment variable first
	tempDir := os.Getenv("TEMP")
	if tempDir == "" {
		tempDir = os.Getenv("TMP")
	}
	if tempDir == "" {
		// Fallback to Windows temp directory
		tempDir = `C:\Windows\Temp`
	}
	return filepath.Join(tempDir, filename)
}

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
		config.DaemonLogFilePath = getDefaultLogPath("daytona-daemon.log")
	}

	if config.EntrypointLogFilePath == "" {
		config.EntrypointLogFilePath = getDefaultLogPath("daytona-entrypoint.log")
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
