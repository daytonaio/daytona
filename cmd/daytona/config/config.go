// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

const configDoesntExistError = "config does not exist. Run `daytona serve` to create a default profile or `daytona profile add` to connect to a remote server."

func IsNotExist(err error) bool {
	return err.Error() == configDoesntExistError
}

func GetConfig() (*Config, error) {
	configFilePath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(configFilePath)
	if os.IsNotExist(err) {
		return nil, errors.New(configDoesntExistError)
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

func (c *Config) GetActiveProfile() (Profile, error) {
	for _, profile := range c.Profiles {
		if profile.Id == c.ActiveProfileId {
			return profile, nil
		}
	}

	return Profile{}, errors.New("active profile not found")
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
