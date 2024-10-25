// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/daytonaio/daytona/pkg/cmd/autocomplete"
	"github.com/google/uuid"
)

type ServerApi struct {
	Url string `json:"url"`
	Key string `json:"key"`
}

type Profile struct {
	Id   string    `json:"id"`
	Name string    `json:"name"`
	Api  ServerApi `json:"api"`
}

type Config struct {
	Id               string    `json:"id"`
	ActiveProfileId  string    `json:"activeProfile"`
	DefaultIdeId     string    `json:"defaultIde"`
	Profiles         []Profile `json:"profiles"`
	TelemetryEnabled bool      `json:"telemetryEnabled"`
}

type Ide struct {
	Id   string
	Name string
}

const DefaultIdeId = "vscode"

type GitProvider struct {
	Id   string
	Name string
}

func GetConfig() (*Config, error) {
	configFilePath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(configFilePath)
	if os.IsNotExist(err) {
		// Setup autocompletion when adding initial config
		_ = autocomplete.DetectShellAndSetupAutocompletion(autocomplete.AutoCompleteCmd.Root())

		config := &Config{
			Id:               uuid.NewString(),
			DefaultIdeId:     DefaultIdeId,
			TelemetryEnabled: true,
		}
		return config, config.Save()
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
		err := c.Save()
		if err != nil {
			return nil, err
		}
	}

	return &c, nil
}

var ErrNoProfilesFound = errors.New("no profiles found. Run `daytona serve` to create a default profile or `daytona profile add` to connect to a remote server")

func (c *Config) GetActiveProfile() (Profile, error) {
	if len(c.Profiles) == 0 {
		return Profile{}, ErrNoProfilesFound
	}

	for _, profile := range c.Profiles {
		if profile.Id == c.ActiveProfileId {
			return profile, nil
		}
	}

	return Profile{}, errors.New("active profile not found. Set an active profile with `daytona profile use`")
}

func (c *Config) Save() error {
	configFilePath, err := getConfigPath()
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(configFilePath), 0755)
	if err != nil {
		return err
	}

	configContent, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFilePath, configContent, 0644)
}

func (c *Config) AddProfile(profile Profile) error {
	c.Profiles = append(c.Profiles, profile)
	c.ActiveProfileId = profile.Id

	return c.Save()
}

func (c *Config) EditProfile(profile Profile) error {
	for i, p := range c.Profiles {
		if p.Id == profile.Id {
			c.Profiles[i] = profile

			return c.Save()
		}
	}

	return fmt.Errorf("profile with id %s not found", profile.Id)
}

func (c *Config) RemoveProfile(profileId string) error {
	if profileId == "default" {
		return errors.New("can not remove default profile")
	}

	var profiles []Profile
	for _, profile := range c.Profiles {
		if profile.Id != profileId {
			profiles = append(profiles, profile)
		}
	}

	if c.ActiveProfileId == profileId {
		c.ActiveProfileId = "default"
	}

	c.Profiles = profiles

	return c.Save()
}

func (c *Config) GetProfile(profileId string) (Profile, error) {
	for _, profile := range c.Profiles {
		if profile.Id == profileId {
			return profile, nil
		}
	}

	return Profile{}, errors.New("profile not found")
}

func (c *Config) EnableTelemetry() error {
	c.TelemetryEnabled = true

	return c.Save()
}

func (c *Config) DisableTelemetry() error {
	c.TelemetryEnabled = false

	return c.Save()
}

func getConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "config.json"), nil
}

func GetConfigDir() (string, error) {
	daytonaConfigDir := os.Getenv("DAYTONA_CONFIG_DIR")
	if daytonaConfigDir != "" {
		return daytonaConfigDir, nil
	}

	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(userConfigDir, "daytona"), nil
}

func DeleteConfigDir() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	err = os.RemoveAll(configDir)
	if err != nil {
		return err
	}

	return nil
}

func TelemetryEnabled() bool {
	telemetryEnabled := os.Getenv("DAYTONA_TELEMETRY_ENABLED")
	if telemetryEnabled != "" {
		return telemetryEnabled == "true"
	}

	c, err := GetConfig()
	if err != nil {
		return false
	}

	return c.TelemetryEnabled
}

func GetClientId() string {
	clientId := os.Getenv("DAYTONA_CLIENT_ID")
	if clientId != "" {
		return clientId
	}

	c, err := GetConfig()
	if err != nil {
		return ""
	}

	return c.Id
}

func GetErrorLogsDir() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	errorLogsDir := filepath.Join(configDir, "error_logs")
	err = os.MkdirAll(errorLogsDir, 0755)
	if err != nil {
		return "", err
	}

	return errorLogsDir, nil
}
