// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

const DEFAULT_POLLER_INTERVAL = "0 */5 * * * *"

type Config struct {
	Id               string `json:"id" validate:"required"`
	Interval         string `json:"interval" validate:"required"`
	TelemetryEnabled bool   `json:"telemetryEnabled" validate:"required"`
} // @name BuildRunnerConfig

func GetConfig() (*Config, error) {
	configFilePath, err := configFilePath()
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(configFilePath)
	if os.IsNotExist(err) {
		c := getDefaultConfig()

		err = Save(*c)
		if err != nil {
			return nil, fmt.Errorf("failed to save default config file: %v", err)
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

func configFilePath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "config.json"), nil
}

func Save(c Config) error {
	configFilePath, err := configFilePath()
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

func GetConfigDir() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(userConfigDir, "daytona", "build-poller"), nil
}

func getDefaultConfig() *Config {
	return &Config{
		Id:               uuid.NewString(),
		Interval:         DEFAULT_POLLER_INTERVAL,
		TelemetryEnabled: false,
	}
}
