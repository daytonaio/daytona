// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"encoding/json"
	"runtime"
	"net/http"	
	"github.com/spf13/cobra"
	"github.com/inconshreveable/go-update"
)

const (
	githubReleaseURL = "https://api.github.com/repos/daytonaio/daytona/releases"
)


type GitHubRelease struct {
	TagName string `json:"tag_name"`
	ChangeLog string `json:"body"`
}

func fetchLatestRelase()(*GitHubRelease, error){
	resp, err := http.Get(githubReleaseURL+"/latest")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch latest release: %s", resp.Status)
	}
	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}
	return &release, nil
}

func fetchVersionRelease(version string)(*GitHubRelease, error){
	resp, err := http.Get(githubReleaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch latest release: %s", resp.Status)
	}
	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}
	for _, release := range releases {
		if release.TagName == version {
			return &release, nil
		}
	}
	return nil, fmt.Errorf("version %s not found", version)
}

var updateCmd = &cobra.Command{
	Use:  "update",
	Short: "Update Daytona CLI",
	Long:  "Update Daytona CLI to the latest version",
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
	},
}
