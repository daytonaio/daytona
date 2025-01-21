// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/google/uuid"

	log "github.com/sirupsen/logrus"
)

type Config struct {
	Id               string              `json:"id"`
	Name             string              `json:"name"`
	ServerApiKey     string              `json:"serverApiKey"`
	ServerApiUrl     string              `json:"serverApiUrl"`
	ProvidersDir     string              `json:"providersDir"`
	LogFile          *logs.LogFileConfig `json:"logFile"`
	ClientId         string              `envconfig:"DAYTONA_CLIENT_ID"`
	TelemetryEnabled bool                `json:"telemetryEnabled"`
} // @name RunnerConfig

var ErrConfigNotFound = errors.New("run 'daytona runner configure' to configure the runner")

func GetConfig() (*Config, error) {
	configFilePath, err := configFilePath()
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(configFilePath)
	if os.IsNotExist(err) {
		return nil, ErrConfigNotFound
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

	if c.ProvidersDir == "" {
		defaultProvidersDir, err := getDefaultProvidersDir()
		if err != nil {
			return nil, err
		}

		c.ProvidersDir = defaultProvidersDir
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

func GetConfigDir() (string, error) {
	daytonaConfigDir := os.Getenv("DAYTONA_RUNNER_CONFIG_DIR")
	if daytonaConfigDir != "" {
		return daytonaConfigDir, nil
	}

	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(userConfigDir, "daytona-runner"), nil
}

func Save(c Config) error {
	if err := util.DirectoryValidator(&c.ProvidersDir); err != nil {
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

func EnableTelemetry(c Config) error {
	c.TelemetryEnabled = true

	return Save(c)
}

func DisableTelemetry(c Config) error {
	c.TelemetryEnabled = false

	return Save(c)
}

func configFilePath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "config.json"), nil
}

func getDefaultLogFilePath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "log"), nil
}

func GetDefaultConfig() (*Config, error) {
	providersDir, err := getDefaultProvidersDir()
	if err != nil {
		return nil, errors.New("failed to get default providers dir")
	}

	logFilePath, err := getDefaultLogFilePath()
	if err != nil {
		log.Error("failed to get default log file path")
	}

	c := Config{
		ProvidersDir: providersDir,
		LogFile:      logs.GetDefaultLogFileConfig(logFilePath),
	}

	if os.Getenv("DEFAULT_PROVIDERS_DIR") != "" {
		c.ProvidersDir = os.Getenv("DEFAULT_PROVIDERS_DIR")
	}

	return &c, nil
}

func getDefaultProvidersDir() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "providers"), nil
}

func GetLogsDir(configDir string) string {
	return filepath.Join(configDir, "logs")
}
