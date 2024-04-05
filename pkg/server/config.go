// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func GetConfig() (*ServerConfig, error) {
	configFilePath, err := configFilePath()
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(configFilePath)
	if os.IsNotExist(err) {
		c, err := getDefaultConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get default config: %v", err)
		}

		err = Save(*c)
		if err != nil {
			return nil, fmt.Errorf("failed to save default config file: %v", err)
		}

		return c, nil
	}

	if err != nil {
		return nil, err
	}

	var c ServerConfig
	configContent, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(configContent, &c)
	if err != nil {
		return nil, err
	}

	if c.BinariesPath == "" {
		binariesPath, err := getDefaultBinariesPath()
		if err != nil {
			return nil, err
		}

		c.BinariesPath = binariesPath
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

func Save(c ServerConfig) error {
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
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(userConfigDir, "daytona", "server"), nil
}

func GetWorkspaceLogsDir() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "logs"), nil
}

func GetWorkspaceLogFilePath(workspaceId string) (string, error) {
	projectLogsDir, err := GetWorkspaceLogsDir()
	if err != nil {
		return "", err
	}

	filePath := filepath.Join(projectLogsDir, workspaceId, "log")

	err = os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

func DeleteWorkspaceLogs(workspaceId string) error {
	logsDir, err := GetWorkspaceLogsDir()
	if err != nil {
		return err
	}

	workspaceLogsDir := filepath.Join(logsDir, workspaceId)

	_, err = os.Stat(workspaceLogsDir)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	return os.RemoveAll(workspaceLogsDir)
}
