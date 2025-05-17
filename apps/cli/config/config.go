// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/daytonaio/daytona/cli/cmd"
	"github.com/daytonaio/daytona/cli/internal"
)

type Config struct {
	ActiveProfileId string    `json:"activeProfile"`
	Profiles        []Profile `json:"profiles"`
}

type Profile struct {
	Id                   string    `json:"id"`
	Name                 string    `json:"name"`
	Api                  ServerApi `json:"api"`
	ActiveOrganizationId *string   `json:"activeOrganizationId"`
}

type ServerApi struct {
	Url   string  `json:"url"`
	Key   *string `json:"key"`
	Token *Token  `json:"token"`
}

type Token struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

func GetConfig() (*Config, error) {
	configFilePath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(configFilePath)
	if os.IsNotExist(err) {
		// Setup autocompletion when adding initial config
		_ = cmd.DetectShellAndSetupAutocompletion(cmd.AutoCompleteCmd.Root())

		config := &Config{}
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

	return &c, nil
}

var ErrNoProfilesFound = errors.New("no profiles found. Run `daytona login` to authenticate")

func (c *Config) GetActiveProfile() (Profile, error) {
	if len(c.Profiles) == 0 {
		return Profile{}, ErrNoProfilesFound
	}

	for _, profile := range c.Profiles {
		if profile.Id == c.ActiveProfileId {
			return profile, nil
		}
	}

	return Profile{}, ErrNoActiveProfile
}

var ErrNoActiveProfile = errors.New("no active profile found. Run `daytona login` to authenticate")
var ErrNoActiveOrganization = errors.New("no active organization found. Run `daytona organization use` to select an organization")

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
	if c.ActiveProfileId == profileId {
		return errors.New("cannot remove active profile")
	}

	var profiles []Profile
	for _, profile := range c.Profiles {
		if profile.Id != profileId {
			profiles = append(profiles, profile)
		}
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

	return os.RemoveAll(configDir)
}

func GetActiveOrganizationId() (string, error) {
	c, err := GetConfig()
	if err != nil {
		return "", err
	}

	activeProfile, err := c.GetActiveProfile()
	if err != nil {
		return "", err
	}

	if activeProfile.ActiveOrganizationId == nil {
		return "", ErrNoActiveOrganization
	}

	return *activeProfile.ActiveOrganizationId, nil
}

func IsApiKeyAuth() bool {
	c, err := GetConfig()
	if err != nil {
		return false
	}

	activeProfile, err := c.GetActiveProfile()
	if err != nil {
		return false
	}

	return activeProfile.Api.Key != nil && activeProfile.Api.Token == nil
}

func GetAuth0Domain() string {
	auth0Domain := os.Getenv("DAYTONA_AUTH0_DOMAIN")
	if auth0Domain == "" {
		auth0Domain = internal.Auth0Domain
	}

	return auth0Domain
}

func GetAuth0ClientId() string {
	auth0ClientId := os.Getenv("DAYTONA_AUTH0_CLIENT_ID")
	if auth0ClientId == "" {
		auth0ClientId = internal.Auth0ClientId
	}

	return auth0ClientId
}

func GetAuth0ClientSecret() string {
	auth0ClientSecret := os.Getenv("DAYTONA_AUTH0_CLIENT_SECRET")
	if auth0ClientSecret == "" {
		auth0ClientSecret = internal.Auth0ClientSecret
	}

	return auth0ClientSecret
}

func GetAuth0CallbackPort() string {
	auth0CallbackPort := os.Getenv("DAYTONA_AUTH0_CALLBACK_PORT")
	if auth0CallbackPort == "" {
		auth0CallbackPort = internal.Auth0CallbackPort
	}

	return auth0CallbackPort
}

func GetAuth0Audience() string {
	auth0Audience := os.Getenv("DAYTONA_AUTH0_AUDIENCE")
	if auth0Audience == "" {
		auth0Audience = internal.Auth0Audience
	}

	return auth0Audience
}

func GetDaytonaApiUrl() string {
	daytonaApiUrl := os.Getenv("DAYTONA_API_URL")
	if daytonaApiUrl == "" {
		daytonaApiUrl = internal.DaytonaApiUrl
	}

	return daytonaApiUrl
}
