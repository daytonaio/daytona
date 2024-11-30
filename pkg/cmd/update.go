// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/inconshreveable/go-update"
	"github.com/spf13/cobra"
	"net/http"
	"runtime"
)

const (
	githubReleaseURL = "https://api.github.com/repos/daytonaio/daytona/releases"
	baseDownloadURL  = "https://download.daytona.io/daytona/"
)

type GitHubRelease struct {
	TagName   string `json:"tag_name"`
	ChangeLog string `json:"body"`
}

func fetchLatestRelase() (*GitHubRelease, error) {
	resp, err := http.Get(githubReleaseURL + "/latest")
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

func fetchVersionRelease(version string) (*GitHubRelease, error) {
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

func updateToVersion(version string) error {
	url := getBinaryUrl(version)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch binary: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch binary: %s", resp.Status)
	}
	err = update.Apply(resp.Body, update.Options{})
	if err != nil {
		fmt.Println("Failed to update binary")
		if rollbackErr := update.RollbackError(err); rollbackErr != nil {
			return fmt.Errorf("failed to rollback: %w", rollbackErr)
		}
		return fmt.Errorf("failed to update binary: %w", err)
	}
	fmt.Println("Successfully updated to version", version)
	return nil
}

func getBinaryUrl(version string) string {
	if runtime.GOOS == "windows" {
		return baseDownloadURL + fmt.Sprintf("%s/daytona-%s-%s.exe", version, runtime.GOOS, runtime.GOARCH)
	}
	return baseDownloadURL + fmt.Sprintf("%s/daytona-%s-%s", version, runtime.GOOS, runtime.GOARCH)
}

var versionFlag string
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Daytona CLI",
	Long:  "Update Daytona CLI to the latest version",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var version string
		var changeLog string
		if versionFlag == "" {
			release, err := fetchLatestRelase()
			if err != nil {
				return err
			}
			version = release.TagName
			changeLog = release.ChangeLog	
		} else {
			release, err := fetchVersionRelease("v" + versionFlag)
			if err != nil {
				return err
			}
			version = release.TagName
			changeLog = release.ChangeLog
		}
		fmt.Println("Updating to version", version)
		fmt.Println("Changelog:")
		fmt.Println(changeLog)
		updateToVersion(version)
		return nil
	},
}

func init() {
	updateCmd.Flags().StringVarP(&versionFlag, "version", "v", "", "Version to update to")
	rootCmd.AddCommand(updateCmd)
}
