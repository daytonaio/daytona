// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"encoding/json"
	"os"
	"path"
)

type ReverseProxyConfig struct {
	Addr string `json:"addr"`
	Port int    `json:"port"`
	Auth string `json:"auth"`
}

type Config struct {
	ReverseProxy ReverseProxyConfig `json:"reverse_proxy"`

	//	WorkspaceDir string `json:"workspace_dir"`
	//	PluginsDir   string `json:"plugins_dir"`
}

func GetConfig() (*Config, error) {
	configFilePath, err := configFilePath()
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(configFilePath)
	if os.IsNotExist(err) {
		panic("config file does not exist")
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
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return path.Join(userConfigDir, "daytona", "agent", "config.json"), nil
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
