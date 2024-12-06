// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/daytonaio/daytona/internal"
	"github.com/inconshreveable/go-update"
	"github.com/spf13/cobra"
	"net/http"
	"runtime"
	"strconv"
	"strings"
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
	if len(releases) == 0 {
		return nil, fmt.Errorf("no releases found")
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
func isVersionGreater(version string, currentVersion string) bool {
	version = strings.TrimPrefix(version, "v")
	currentVersion = strings.TrimPrefix(currentVersion, "v")
	v1 := strings.Split(version, ".")
	v2 := strings.Split(currentVersion, ".")
	maxLength := len(v1)
	if len(v2) > maxLength {
		maxLength = len(v2)
	}
	for i := 0; i < maxLength; i++ {
		var ver1 int
		var ver2 int
		if i < len(v1) {
			ver1, _ = strconv.Atoi(v1[i])
		}
		if i < len(v2) {
			ver2, _ = strconv.Atoi(v2[i])
		}
		if ver1 > ver2 {
			return false
		}
		if ver1 < ver2 {
			return true
		}
	}
	return false
}
func isCurrentVersionEqualToPrevious(currentVersion string, version string) (bool, error) {
	resp, err := http.Get(githubReleaseURL)
	if err != nil {
		return false, fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("failed to fetch latest release: %s", resp.Status)
	}
	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return false, fmt.Errorf("failed to fetch latest release: %w", err)
	}
	if len(releases) == 0 {
		return false, fmt.Errorf("no releases found")
	}
	for i := 0; i < len(releases); i++ {
		if releases[i].TagName == version {
			if i+1 < len(releases) {
				if releases[i+1].TagName == currentVersion {
					return true, nil
				}
			}
		}
	}
	return false, nil
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
		currentVersion := internal.Version
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
		if isVersionGreater(version, currentVersion) {
			fmt.Println("Current version is greater than the version you are trying to update to")
			return nil
		}
		isCurrEqualPrev, err := isCurrentVersionEqualToPrevious(currentVersion, version)
		if err != nil {
			return err
		}
		if !isCurrEqualPrev {
			changeLog += "\nThere are more changes in the version. Please visit https://github.com/daytonaio/daytona/releases for all the complete chnageLog"
		}
		fmt.Println("Updating to version ", version, "from ", currentVersion)
		fmt.Println("ChangeLog:")
		fmt.Println(changeLog)
		err = updateToVersion(version)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	updateCmd.Flags().StringVarP(&versionFlag, "version", "v", "", "Version to update to")
	rootCmd.AddCommand(updateCmd)
}
