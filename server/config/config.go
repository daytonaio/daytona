// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"encoding/json"
	"os"
	"path"
)

type Config struct {
	DefaultWorkspaceDir string `json:"defaultWorkspaceDir"`
	ProjectBaseImage    string `json:"projectBaseImage"`
	PluginsDir          string `json:"pluginsDir"`
}

func GetConfig() (*Config, error) {
	configFilePath, err := configFilePath()
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(configFilePath)
	if os.IsNotExist(err) {
		defaultWorkspaceDir, err := getDefaultWorkspaceDir()
		if err != nil {
			return nil, err
		}
		pluginsDir, err := getDefaultPluginsDir()
		if err != nil {
			return nil, err
		}

		return &Config{
			DefaultWorkspaceDir: defaultWorkspaceDir,
			ProjectBaseImage:    defaultProjectBaseImage,
			PluginsDir:          pluginsDir,
		}, nil
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

	return &c, nil
}

func configFilePath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	return path.Join(configDir, "config.json"), nil
}

func (c *Config) Save() error {
	configFilePath, err := configFilePath()
	if err != nil {
		return err
	}

	configContent, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	err = os.MkdirAll(path.Dir(configFilePath), 0700)
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

	return path.Join(userConfigDir, "daytona", "server"), nil
}

func getDefaultWorkspaceDir() (string, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	defaultWorkspaceDir := path.Join(userHomeDir, ".daytona_workspaces")
	err = os.MkdirAll(defaultWorkspaceDir, 0700)
	if err != nil {
		return "", err
	}

	return defaultWorkspaceDir, nil
}

func getDefaultPluginsDir() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return path.Join(userConfigDir, "daytona", "plugins"), nil
}
