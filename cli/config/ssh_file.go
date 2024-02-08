// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
)

func ensureSshFilesLinked() error {
	// Make sure ~/.ssh/config file exists
	sshDir := path.Join(os.Getenv("HOME"), ".ssh")
	configPath := path.Join(sshDir, "config")

	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		err := os.WriteFile(configPath, []byte{}, 0600)
		if err != nil {
			return err
		}
	}

	// Make sure daytona_config file exists
	daytonaConfigPath := path.Join(sshDir, "daytona_config")

	_, err = os.Stat(daytonaConfigPath)
	if os.IsNotExist(err) {
		err := os.WriteFile(daytonaConfigPath, []byte{}, 0600)
		if err != nil {
			return err
		}
	}

	// Make sure daytona_config is included
	configFile := path.Join(sshDir, "config")
	_, err = os.Stat(configFile)
	if os.IsNotExist(err) {
		err := os.WriteFile(configFile, []byte("Include daytona_config\n\n"), 0600)
		if err != nil {
			return err
		}
	} else {
		content, err := os.ReadFile(configFile)
		if err != nil {
			return err
		}

		if strings.Contains(string(content), "Include daytona_config") {
			return nil
		}

		newContent := "Include daytona_config\n\n" + string(content)
		err = os.WriteFile(configFile, []byte(newContent), 0600)
		if err != nil {
			return err
		}
	}

	return nil
}

// Add ssh entry

func generateSshConfigEntry(profileId, workspaceName, projectName, knownHostsPath string) (string, error) {
	daytonaPath, err := os.Executable()
	if err != nil {
		return "", err
	}

	tab := "\t"
	projectHostname := GetProjectHostname(profileId, workspaceName, projectName)

	config := fmt.Sprintf("Host %s\n"+
		tab+"User %s~%s\n"+
		tab+"StrictHostKeyChecking no\n"+
		tab+"UserKnownHostsFile %s\n"+
		tab+"ProxyCommand %s ssh-proxy %s %s %s\n\n", projectHostname, workspaceName, projectName, knownHostsPath, daytonaPath, profileId, workspaceName, projectName)

	return config, nil
}

func EnsureSshConfigEntryAdded(profileId, workspaceName, projectName string) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	err = ensureSshFilesLinked()
	if err != nil {
		return err
	}

	data, err := generateSshConfigEntry(profileId, workspaceName, projectName, path.Join(configDir, ".known_hosts"))
	if err != nil {
		return err
	}

	// Make sure ~/.ssh/config file exists
	sshDir := path.Join(os.Getenv("HOME"), ".ssh")
	configPath := path.Join(sshDir, "daytona_config")

	// Read existing content from the file
	existingContent, err := os.ReadFile(configPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if strings.Contains(string(existingContent), data) {
		return nil
	}

	// Combine the new data with existing content
	newData := data + string(existingContent)

	// Open the file for writing
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(newData)
	if err != nil {
		return err
	}

	return nil
}

func RemoveWorkspaceSshEntries(profileId, workspaceName string) error {
	sshDir := path.Join(os.Getenv("HOME"), ".ssh")
	configPath := path.Join(sshDir, "daytona_config")

	// Read existing content from the file
	existingContent, err := os.ReadFile(configPath)
	if err != nil && !os.IsNotExist(err) {
		return nil
	}

	regex := regexp.MustCompile(fmt.Sprintf(`Host %s~%s~\w+\n(?:\t.*\n?)*`, profileId, workspaceName))
	newContent := regex.ReplaceAllString(string(existingContent), "")

	newContent = strings.Trim(newContent, "\n")

	// Open the file for writing
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(newContent)
	if err != nil {
		return err
	}

	return nil
}

func GetProjectHostname(profileId, workspaceName, projectName string) string {
	return fmt.Sprintf("%s~%s~%s", profileId, workspaceName, projectName)
}
