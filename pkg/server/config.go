// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func GetConfig() (*Config, error) {
	configFilePath, err := configFilePath()
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(configFilePath)
	if os.IsNotExist(err) {
		c, err := getDefaultConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get default config: %w", err)
		}

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

	if c.LogFile == nil {
		logFilePath, err := getDefaultLogFilePath()
		if err != nil {
			log.Error("failed to get default log file path")
		}

		c.LogFile = logs.GetDefaultLogFileConfig(logFilePath)
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
	if err := util.DirectoryValidator(&c.BinariesPath); err != nil {
		return err
	}

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

	err = os.WriteFile(configFilePath, configContent, 0600)
	if err != nil {
		return err
	}

	return nil
}

func GetConfigDir() (string, error) {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "server"), nil
}

func GetTargetLogsDir(configDir string) string {
	return filepath.Join(configDir, "targets", "logs")
}

func GetRunnerLogsDir(configDir string) string {
	return filepath.Join(configDir, "runners", "logs")
}

func GetWorkspaceLogsDir(configDir string) string {
	return filepath.Join(configDir, "workspaces", "logs")
}
