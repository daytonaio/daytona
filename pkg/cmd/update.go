// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/daytonaio/daytona/internal"
	"github.com/inconshreveable/go-update"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
	"net/http"
	"net/url"
	"regexp"
	"runtime"
	"strings"
)

const (
	githubReleaseURL = "https://api.github.com/repos/daytonaio/daytona/releases"
	baseDownloadURL  = "https://download.daytona.io/daytona/"
)

type GitHubRelease struct {
	TagName   string `json:"tag_name"`
	Changelog string `json:"body"`
}

var versionFlag string
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Daytona CLI",
	Long:  "Update Daytona CLI to the latest version",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var version string
		var changelog string
		currentVersion := internal.Version
		if versionFlag == "" {
			release, err := fetchLatestRelase()
			if err != nil {
				return err
			}
			version = release.TagName
			changelog = release.Changelog
		} else {
			semverRegex := `^v?(\d+\.\d+\.\d+)$`
			matched, err := regexp.MatchString(semverRegex, versionFlag)
			if err != nil {
				return err
			}
			if !matched {
				return errors.New("invalid version format: expected 'vX.Y.X' or 'X.Y.Z'")
			}
			if !strings.HasPrefix(versionFlag, "v") {
				versionFlag = "v" + versionFlag
			}
			release, err := fetchVersionRelease(versionFlag)
			if err != nil {
				return err
			}
			version = release.TagName
			changelog = release.Changelog
		}
		isCurrVerEqToPrev, err := isCurrentVersionEqualToPrevious(currentVersion, version)
		if err != nil {
			return err
		}

		if isCurrVerEqToPrev {
			fmt.Println("Current version is greater than the version you are trying to update to")
			return nil
		}
		if semver.Compare(currentVersion, version) != 0 {
			changelog += "\n\nThere might be more important changes since you updated. Please visit https://github.com/daytonaio/daytona/releases for the complete changelog\n"
		}
		fmt.Println("Updating to version", version, "from", currentVersion)
		fmt.Println("\nChangelog:")
		fmt.Println(changelog)
		return updateToVersion(version)

	},
}

func init() {
	updateCmd.Flags().StringVarP(&versionFlag, "version", "v", "", "Version to update to")

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
	url, err := getBinaryUrl(version)
	if err != nil {
		return fmt.Errorf("failed to get binary url: %w", err)
	}
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
	fmt.Println("\nSuccessfully updated to version", version)
	return nil
}

func getBinaryUrl(version string) (string, error) {
	fileName := fmt.Sprintf("daytona-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		fileName += ".exe"
	}
	fullURL, err := url.JoinPath(baseDownloadURL, version, fileName)
	if err != nil {
		return "", err
	}
	return fullURL, nil
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
