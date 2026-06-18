// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package config_test

import (
	"errors"
	"testing"

	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/cli/internal"
)

func TestGetActiveProfile(t *testing.T) {
	storedKey := "stored-key"
	cfg := config.Config{
		ActiveProfileId: "p1",
		Profiles: []config.Profile{
			{Id: "p1", Name: "default", Api: config.ServerApi{Url: "https://stored.example.com/api", Key: &storedKey}},
		},
	}

	tests := []struct {
		name       string
		envKey     string
		envUrl     string
		builtinUrl string
		wantId     string
		wantUrl    string
	}{
		{
			name:    "key and url env set uses env profile",
			envKey:  "env-key",
			envUrl:  "https://env.example.com/api",
			wantId:  "env",
			wantUrl: "https://env.example.com/api",
		},
		{
			name:       "key env only falls back to built-in api url",
			envKey:     "env-key",
			builtinUrl: "https://builtin.example.com/api",
			wantId:     "env",
			wantUrl:    "https://builtin.example.com/api",
		},
		{
			name:    "key env only without built-in url falls back to stored profile",
			envKey:  "env-key",
			wantId:  "p1",
			wantUrl: "https://stored.example.com/api",
		},
		{
			name:    "url env only falls back to stored profile",
			envUrl:  "https://env.example.com/api",
			wantId:  "p1",
			wantUrl: "https://stored.example.com/api",
		},
		{
			name:    "no env uses active profile",
			wantId:  "p1",
			wantUrl: "https://stored.example.com/api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(config.DAYTONA_API_KEY_ENV_VAR, tt.envKey)
			t.Setenv(config.DAYTONA_API_URL_ENV_VAR, tt.envUrl)
			prevBuiltin := internal.DaytonaApiUrl
			internal.DaytonaApiUrl = tt.builtinUrl
			t.Cleanup(func() { internal.DaytonaApiUrl = prevBuiltin })

			profile, err := cfg.GetActiveProfile()
			if err != nil {
				t.Fatalf("GetActiveProfile() unexpected error: %v", err)
			}
			if profile.Id != tt.wantId {
				t.Errorf("GetActiveProfile() profile id = %q, want %q", profile.Id, tt.wantId)
			}
			if profile.Api.Url != tt.wantUrl {
				t.Errorf("GetActiveProfile() api url = %q, want %q", profile.Api.Url, tt.wantUrl)
			}
		})
	}
}

func TestGetActiveProfileNoProfiles(t *testing.T) {
	t.Setenv(config.DAYTONA_API_KEY_ENV_VAR, "")
	t.Setenv(config.DAYTONA_API_URL_ENV_VAR, "")

	cfg := config.Config{}
	_, err := cfg.GetActiveProfile()
	if !errors.Is(err, config.ErrNoProfilesFound) {
		t.Errorf("GetActiveProfile() error = %v, want ErrNoProfilesFound", err)
	}
}
