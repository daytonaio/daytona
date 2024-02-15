// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"path"

	"github.com/daytonaio/daytona/common/grpc/proto/types"
	"github.com/google/uuid"
)

func GetConfig() (*types.ServerConfig, error) {
	configFilePath, err := configFilePath()
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(configFilePath)
	if os.IsNotExist(err) {
		return nil, errors.New("config file does not exist")
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

func generateUuid() string {
	uuid := uuid.New()
	return uuid.String()
}

func init() {
	_, err := GetConfig()
	if err == nil {
		return
	}

	defaultWorkspaceDir, err := getDefaultWorkspaceDir()
	if err != nil {
		log.Fatal("failed to get default workspace dir")
	}
	pluginsDir, err := getDefaultPluginsDir()
	if err != nil {
		log.Fatal("failed to get default plugins dir")
	}

	c := types.ServerConfig{
		DefaultWorkspaceDir: defaultWorkspaceDir,
		ProjectBaseImage:    defaultProjectBaseImage,
		PluginRegistryUrl:   defaultPluginRegistryUrl,
		PluginsDir:          pluginsDir,
		ServerDownloadUrl:   defaultServerDownloadUrl,
		GrpcPort:            defaultGrpcPort,
		HeadscalePort:       defaultHeadscalePort,
		Frps: &types.FRPSConfig{
			Domain:   defaultFrpsDomain,
			Port:     defaultFrpsPort,
			Protocol: defaultFrpsProtocol,
		},
		Id: generateUuid(),
	}

	err = Save(&c)
	if err != nil {
		log.Fatal("failed to save default config file")
	}
}
