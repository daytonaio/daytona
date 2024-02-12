// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"encoding/json"
	"os"
	"path"

	"github.com/daytonaio/daytona/common/grpc/proto/types"
)

const workspaceKeyFileName = "workspace_key"
const defaultProjectBaseImage = "daytonaio/workspace-project:latest"
const defaultPluginRegistryUrl = "https://download.daytona.io/daytona/plugins"

func GetConfig() (*types.ServerConfig, error) {
	configFilePath, err := configFilePath()
	if err != nil {
		return nil, err
	}

	defaultWorkspaceDir, err := getDefaultWorkspaceDir()
	if err != nil {
		return nil, err
	}
	pluginsDir, err := getDefaultPluginsDir()
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(configFilePath)
	if os.IsNotExist(err) {
		return &types.ServerConfig{
			DefaultWorkspaceDir: defaultWorkspaceDir,
			ProjectBaseImage:    defaultProjectBaseImage,
			PluginRegistryUrl:   defaultPluginRegistryUrl,
			PluginsDir:          pluginsDir,
		}, nil
	}

	if err != nil {
		return nil, err
	}

	var c types.ServerConfig
	configContent, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(configContent, &c)
	if err != nil {
		return nil, err
	}

	if c.DefaultWorkspaceDir == "" {
		c.DefaultWorkspaceDir = defaultWorkspaceDir
	}
	if c.ProjectBaseImage == "" {
		c.ProjectBaseImage = defaultProjectBaseImage
	}
	if c.PluginRegistryUrl == "" {
		c.PluginRegistryUrl = defaultPluginRegistryUrl
	}
	if c.PluginsDir == "" {
		c.PluginsDir = pluginsDir
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

func Save(c *types.ServerConfig) error {
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
