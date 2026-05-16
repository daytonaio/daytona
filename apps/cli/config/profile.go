// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/daytonaio/daytona/cli/internal"
	"github.com/iancoleman/strcase"
)

const DaytonaApiUrlEnvVar = "DAYTONA_API_URL"
const DaytonaApiKeyEnvVar = "DAYTONA_API_KEY"

type Profile struct {
	Id                   string            `json:"id"`
	Name                 string            `json:"name"`
	Api                  ServerApi         `json:"api"`
	ActiveOrganizationId *string           `json:"activeOrganizationId"`
	ToolboxProxyUrls     map[string]string `json:"toolboxProxyUrls,omitempty"` // Cache proxy URLs by region
}

func (c *Config) GetActiveProfile() (Profile, error) {
	apiUrl := os.Getenv(DaytonaApiUrlEnvVar)
	apiKey := os.Getenv(DaytonaApiKeyEnvVar)

	if apiUrl != "" && apiKey != "" {
		return Profile{
			Id: "env",
			Api: ServerApi{
				Url: apiUrl,
				Key: &apiKey,
			},
		}, nil
	}

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

func (c *Config) ProfileExists(profileId string) bool {
	for _, profile := range c.Profiles {
		if profile.Id == profileId {
			return true
		}
	}

	return false
}

func CreateProfile(
	profileName string,
	c *Config,
) (Profile, error) {
	profile := Profile{
		Id:   strcase.ToSnake(profileName),
		Name: profileName,
		Api: ServerApi{
			Url: GetDaytonaApiUrl(),
		},
	}

	if internal.Version == "v0.0.0-dev" {
		profile.Api.Url = "http://localhost:3001/api"
	}

	return profile, c.AddProfile(profile)
}
