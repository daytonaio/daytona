// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_ProfileManagement(t *testing.T) {
	// Create a temporary config directory
	tempDir, err := os.MkdirTemp("", "daytona-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Set the config directory environment variable
	originalConfigDir := os.Getenv("DAYTONA_CONFIG_DIR")
	defer os.Setenv("DAYTONA_CONFIG_DIR", originalConfigDir)
	os.Setenv("DAYTONA_CONFIG_DIR", tempDir)

	// Test creating a new config
	config, err := GetConfig()
	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Empty(t, config.Profiles)
	assert.Empty(t, config.ActiveProfileId)

	// Test adding a profile
	profile1 := Profile{
		Id:   "profile1",
		Name: "Test Profile 1",
		Api: ServerApi{
			Url: "http://localhost:3001/api",
		},
	}

	err = config.AddProfile(profile1)
	require.NoError(t, err)
	assert.Len(t, config.Profiles, 1)
	assert.Equal(t, "profile1", config.ActiveProfileId)

	// Test getting active profile
	activeProfile, err := config.GetActiveProfile()
	require.NoError(t, err)
	assert.Equal(t, "profile1", activeProfile.Id)
	assert.Equal(t, "Test Profile 1", activeProfile.Name)

	// Test adding another profile
	profile2 := Profile{
		Id:   "profile2",
		Name: "Test Profile 2",
		Api: ServerApi{
			Url: "http://localhost:3002/api",
		},
	}

	err = config.AddProfile(profile2)
	require.NoError(t, err)
	assert.Len(t, config.Profiles, 2)
	assert.Equal(t, "profile2", config.ActiveProfileId)

	// Test setting active profile
	err = config.SetActiveProfile("profile1")
	require.NoError(t, err)
	assert.Equal(t, "profile1", config.ActiveProfileId)

	activeProfile, err = config.GetActiveProfile()
	require.NoError(t, err)
	assert.Equal(t, "profile1", activeProfile.Id)

	// Test getting a specific profile
	profile, err := config.GetProfile("profile2")
	require.NoError(t, err)
	assert.Equal(t, "profile2", profile.Id)

	// Test editing a profile
	profile1.Name = "Updated Profile 1"
	err = config.EditProfile(profile1)
	require.NoError(t, err)

	updatedProfile, err := config.GetProfile("profile1")
	require.NoError(t, err)
	assert.Equal(t, "Updated Profile 1", updatedProfile.Name)

	// Test removing a profile (non-active)
	err = config.RemoveProfile("profile2")
	require.NoError(t, err)
	assert.Len(t, config.Profiles, 1)

	// Test removing active profile should fail
	err = config.RemoveProfile("profile1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot remove active profile")

	// Test setting active profile to non-existent profile
	err = config.SetActiveProfile("nonexistent")
	assert.Error(t, err)
}

func TestConfig_BackwardCompatibility(t *testing.T) {
	// Create a temporary config directory
	tempDir, err := os.MkdirTemp("", "daytona-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Set the config directory environment variable
	originalConfigDir := os.Getenv("DAYTONA_CONFIG_DIR")
	defer os.Setenv("DAYTONA_CONFIG_DIR", originalConfigDir)
	os.Setenv("DAYTONA_CONFIG_DIR", tempDir)

	// Create a config file manually with profiles but no active profile
	configDir := filepath.Join(tempDir)
	err = os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	configFilePath := filepath.Join(configDir, "config.json")
	configContent := `{
  "activeProfile": "",
  "profiles": [
    {
      "id": "profile1",
      "name": "Default Profile",
      "api": {
        "url": "http://localhost:3001/api"
      }
    }
  ]
}`

	err = os.WriteFile(configFilePath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Test that GetActiveProfile automatically sets the first profile as active
	config, err := GetConfig()
	require.NoError(t, err)

	activeProfile, err := config.GetActiveProfile()
	require.NoError(t, err)
	assert.Equal(t, "profile1", activeProfile.Id)
	assert.Equal(t, "profile1", config.ActiveProfileId)

	// Verify the config was saved
	config2, err := GetConfig()
	require.NoError(t, err)
	assert.Equal(t, "profile1", config2.ActiveProfileId)
}

func TestConfig_GetProfile_NotFound(t *testing.T) {
	// Create a temporary config directory
	tempDir, err := os.MkdirTemp("", "daytona-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Set the config directory environment variable
	originalConfigDir := os.Getenv("DAYTONA_CONFIG_DIR")
	defer os.Setenv("DAYTONA_CONFIG_DIR", originalConfigDir)
	os.Setenv("DAYTONA_CONFIG_DIR", tempDir)

	config, err := GetConfig()
	require.NoError(t, err)

	_, err = config.GetProfile("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "profile not found")
}

func TestConfig_GetActiveProfile_NoProfiles(t *testing.T) {
	// Create a temporary config directory
	tempDir, err := os.MkdirTemp("", "daytona-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Set the config directory environment variable
	originalConfigDir := os.Getenv("DAYTONA_CONFIG_DIR")
	defer os.Setenv("DAYTONA_CONFIG_DIR", originalConfigDir)
	os.Setenv("DAYTONA_CONFIG_DIR", tempDir)

	config, err := GetConfig()
	require.NoError(t, err)

	_, err = config.GetActiveProfile()
	assert.Error(t, err)
	assert.Equal(t, ErrNoProfilesFound, err)
}
