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
	"regexp"
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
func getChangeLog(currentVersion string, version string) ([]string, error) {
	var changeLog []string
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
	counter := 10
	canAdd := false
	for _, release := range releases {
		if counter == 0 {
			break
		}
		if release.TagName == version {
			canAdd = true
		}
		if canAdd {
			changeLog = append(changeLog, release.ChangeLog)
			counter--
		}
	}
	return changeLog, nil
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

var versionFlag string
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Daytona CLI",
	Long:  "Update Daytona CLI to the latest version",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var version string
		currentVersion := internal.Version
		if versionFlag == "" {
			release, err := fetchLatestRelase()
			if err != nil {
				return err
			}
			version = release.TagName
		} else {
			release, err := fetchVersionRelease("v" + versionFlag)
			if err != nil {
				return err
			}
			version = release.TagName
		}
		if isVersionGreater(version, currentVersion) {
			fmt.Println("Current version is greater than the version you are trying to update to")
			return nil
		}
		changeLog, err := getChangeLog(currentVersion, version)
		if err != nil {
			return err
		}
		fmt.Println("Updating to version", version)
		renderChangeLog(changeLog, version, currentVersion)
		err = updateToVersion(version)
		if err != nil {
			return err
		}
		return nil
	},
}

func renderChangeLog(changeLog []string, version string, currentVersion string) {
	categories := map[string][]string{
		"features": {},
		"fixes":    {},
		"others":   {},
	}

	var changeLogs []string
	for _, log := range changeLog {

		// Match "others" sections
		otherSectionsRegex := regexp.MustCompile(`(?mi)^### ([^\n]+)\n([\s\S]*?)(\n###|\n\*\*|$)`)
		allMatches := otherSectionsRegex.FindAllStringSubmatch(log, -1)
		for _, matches := range allMatches {
			sectionTitle := strings.ToLower(matches[1])
			if sectionTitle != "fixes" && sectionTitle != "features" {
				lines := strings.Split(matches[2], "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line != "" && strings.HasPrefix(line, "*") || strings.Contains(line, "update") {
						changeLogs = append(changeLogs, line)
					}
				}
			}
		}
	}
	for _, log := range changeLogs {
		if strings.HasPrefix(log, "* feat") {
			categories["features"] = append(categories["features"], log)
		} else if strings.HasPrefix(log, "* fix") {
			categories["fixes"] = append(categories["fixes"], log)
		} else {
			categories["others"] = append(categories["others"], log)
		}
	}

	// Display ChangeLog Summary
	if len(changeLog) == 10 {
		fmt.Println("Showing the ChangeLog for the last 10 releases.\nYou can find the complete changeLog at https://github.com/daytonaio/daytona/releases\n")
	} else {
		fmt.Printf("Showing the ChangeLog from %s to %s.\n\n", currentVersion, version)
	}
	if len(categories["features"]) > 0 {
		fmt.Println("Features:")
		for _, feature := range categories["features"] {
			fmt.Printf(" - %s\n", feature)
		}
	}
	if len(categories["fixes"]) > 0 {
		fmt.Println("Bug Fixes:")
		for _, fix := range categories["fixes"] {
			fmt.Printf(" - %s\n", fix)
		}
	}
	if len(categories["others"]) > 0 {
		fmt.Println("Other Improvements:")
		for _, improvement := range categories["others"] {
			fmt.Printf(" - %s\n", improvement)
		}
	}
}
func init() {
	updateCmd.Flags().StringVarP(&versionFlag, "version", "v", "", "Version to update to")
	rootCmd.AddCommand(updateCmd)
}
