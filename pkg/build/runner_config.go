// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/google/uuid"
)

// TODO: add lock when running interval func
// 10 second interval
const DEFAULT_BUILD_POLL_INTERVAL = "*/10 * * * * *"

type Config struct {
	Id               string `json:"id" validate:"required"`
	Interval         string `json:"interval" validate:"required"`
	TelemetryEnabled bool   `json:"telemetryEnabled" validate:"required"`
} // @name BuildRunnerConfig

func GetConfig() (*Config, error) {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(configFilePath)
	if os.IsNotExist(err) {
		c := getDefaultConfig()

		err = Save(*c)
		if err != nil {
			return nil, fmt.Errorf("failed to save default config file: %w", err)
		}

		return c, nil
	}

	if err != nil {
		return nil, err
	}

	var c Config
	configContent, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(configContent, &c)
	if err != nil {
		return nil, err
	}

	if c.Id == "" {
		c.Id = uuid.NewString()
	}
	err = Save(c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func getConfigFilePath() (string, error) {
	configDir, err := GetRunnerConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "config.json"), nil
}

func Save(c Config) error {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	configContent, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(configFilePath), 0700)
	if err != nil {
		return err
	}

	return os.WriteFile(configFilePath, configContent, 0600)
}

func GetRunnerConfigDir() (string, error) {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "build-runner"), nil
}

func getDefaultConfig() *Config {
	return &Config{
		Id:               uuid.NewString(),
		Interval:         DEFAULT_BUILD_POLL_INTERVAL,
		TelemetryEnabled: false,
	}
}
