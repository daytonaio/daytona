// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package config

import (
	"encoding/json"
	"errors"
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

func GetToolboxProxyUrl(region string) (string, error) {
	c, err := GetConfig()
	if err != nil {
		return "", err
	}

	activeProfile, err := c.GetActiveProfile()
	if err != nil {
		return "", err
	}

	if activeProfile.ToolboxProxyUrls == nil {
		return "", nil
	}

	return activeProfile.ToolboxProxyUrls[region], nil
}

func SetToolboxProxyUrl(region, url string) error {
	c, err := GetConfig()
	if err != nil {
		return err
	}

	// Find and update the active profile
	for i, profile := range c.Profiles {
		if profile.Id == c.ActiveProfileId {
			if c.Profiles[i].ToolboxProxyUrls == nil {
				c.Profiles[i].ToolboxProxyUrls = make(map[string]string)
			}
			c.Profiles[i].ToolboxProxyUrls[region] = url
			return c.Save()
		}
	}

	return ErrNoActiveProfile
}
