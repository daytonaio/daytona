// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

	// Backward compatibility: if no active profile is set but profiles exist, use the first one
	if c.ActiveProfileId == "" {
		c.ActiveProfileId = c.Profiles[0].Id
		if err := c.Save(); err != nil {
			return Profile{}, err
		}
		return c.Profiles[0], nil
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

func (c *Config) SetActiveProfile(profileId string) error {
	_, err := c.GetProfile(profileId)
	if err != nil {
		return err
	}

	c.ActiveProfileId = profileId
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

func GetDaytonaApiUrl() string {
	daytonaApiUrl := os.Getenv("DAYTONA_API_URL")
	if daytonaApiUrl == "" {
		daytonaApiUrl = internal.DaytonaApiUrl
	}

	return daytonaApiUrl
}

// OidcConfig represents the OIDC configuration from the API
type OidcConfig struct {
	Issuer   string    `json:"issuer"`
	ClientId string    `json:"clientId"`
	Audience string    `json:"audience"`
	Cli      CliConfig `json:"cli"`
}

// CliConfig represents the CLI-specific configuration from the API
type CliConfig struct {
	ClientId     string `json:"clientId"`
	CallbackPort string `json:"callbackPort"`
}

// DaytonaConfiguration represents the full configuration from the API
type DaytonaConfiguration struct {
	Oidc OidcConfig `json:"oidc"`
}

// CliAuthConfig represents the complete authentication configuration for the CLI
type CliAuthConfig struct {
	Issuer       string
	ClientId     string
	Audience     string
	CallbackPort string
}

// GetCliAuthConfigFromAPI fetches CLI authentication configuration from the Daytona API
// Endpoint: GET {apiUrl}/api/config
// Returns: issuer (from oidc), clientId and callbackPort (from cli), and audience (from oidc)
// This is used for public clients that don't have a client secret
func GetCliAuthConfigFromAPI(apiUrl string) (*CliAuthConfig, error) {
	// Normalize API URL
	apiUrl = strings.TrimSuffix(apiUrl, "/")
	if !strings.HasSuffix(apiUrl, "/api") {
		apiUrl = apiUrl + "/api"
	}

	// Fetch configuration from GET /api/config endpoint
	configUrl := apiUrl + "/config"
	resp, err := http.Get(configUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch CLI auth config from API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch CLI auth config: API returned status %d", resp.StatusCode)
	}

	var config DaytonaConfiguration
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode CLI auth config: %w", err)
	}

	// Combine OIDC config (issuer, audience) with CLI config (clientId, callbackPort)
	return &CliAuthConfig{
		Issuer:       config.Oidc.Issuer,
		ClientId:     config.Oidc.Cli.ClientId,
		Audience:     config.Oidc.Audience,
		CallbackPort: config.Oidc.Cli.CallbackPort,
	}, nil
}

// GetOidcConfigFromAPI is deprecated - use GetCliAuthConfigFromAPI instead
// Kept for backward compatibility
func GetOidcConfigFromAPI(apiUrl string) (*OidcConfig, error) {
	cliConfig, err := GetCliAuthConfigFromAPI(apiUrl)
	if err != nil {
		return nil, err
	}
	return &OidcConfig{
		Issuer:   cliConfig.Issuer,
		ClientId: cliConfig.ClientId,
		Audience: cliConfig.Audience,
	}, nil
}
