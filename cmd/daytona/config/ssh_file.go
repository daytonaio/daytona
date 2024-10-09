// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
)

var sshHomeDir string

func ensureSshFilesLinked() error {
	// Make sure ~/.ssh/config file exists if not create it
	sshDir := filepath.Join(sshHomeDir, ".ssh")
	configPath := filepath.Join(sshDir, "config")

	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		err := os.MkdirAll(sshDir, 0700)
		if err != nil {
			return err
		}
		err = os.WriteFile(configPath, []byte{}, 0600)
		if err != nil {
			return err
		}
	}

	// Make sure daytona_config file exists
	daytonaConfigPath := filepath.Join(sshDir, "daytona_config")

	_, err = os.Stat(daytonaConfigPath)
	if os.IsNotExist(err) {
		err := os.WriteFile(daytonaConfigPath, []byte{}, 0600)
		if err != nil {
			return err
		}
	}

	// Make sure daytona_config is included
	configFile := filepath.Join(sshDir, "config")
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

		newContent := strings.ReplaceAll(string(content), "Include daytona_config\n\n", "")
		newContent = strings.ReplaceAll(string(newContent), "Include daytona_config\n", "")
		newContent = strings.ReplaceAll(string(newContent), "Include daytona_config", "")
		newContent = "Include daytona_config\n\n" + newContent
		err = os.WriteFile(configFile, []byte(newContent), 0600)
		if err != nil {
			return err
		}
	}

	return nil
}

func UnlinkSshFiles() error {
	sshDirPath := filepath.Join(sshHomeDir, ".ssh")
	sshConfigPath := filepath.Join(sshDirPath, "config")
	daytonaConfigPath := filepath.Join(sshDirPath, "daytona_config")

	// Remove the include line from the config file
	_, err := os.Stat(sshConfigPath)
	if os.IsExist(err) {
		content, err := os.ReadFile(sshConfigPath)
		if err != nil {
			return err
		}

		newContent := strings.ReplaceAll(string(content), "Include daytona_config\n\n", "")
		newContent = strings.ReplaceAll(string(newContent), "Include daytona_config", "")
		err = os.WriteFile(sshConfigPath, []byte(newContent), 0600)
		if err != nil {
			return err
		}
	}

	// Remove the daytona_config file
	_, err = os.Stat(daytonaConfigPath)
	if os.IsExist(err) {
		err = os.Remove(daytonaConfigPath)
		if err != nil {
			return err
		}
	}

	return nil
}

// Add ssh entry

func generateSshConfigEntry(profileId, workspaceId, projectName, knownHostsPath string, gpgForward bool) (string, error) {
	daytonaPath, err := os.Executable()
	if err != nil {
		return "", err
	}

	tab := "\t"
	projectHostname := GetProjectHostname(profileId, workspaceId, projectName)

	config := fmt.Sprintf("Host %s\n"+
		tab+"User daytona\n"+
		tab+"StrictHostKeyChecking no\n"+
		tab+"UserKnownHostsFile %s\n"+
		tab+"ProxyCommand \"%s\" ssh-proxy %s %s %s\n"+
		tab+"ForwardAgent yes\n", projectHostname, knownHostsPath, daytonaPath, profileId, workspaceId, projectName)

	if gpgForward {
		localSocket, err := getLocalGPGSocket()
		if err != nil {
			log.Warn(err)
			return config, nil
		}

		remoteSocket, err := getRemoteGPGSocket(projectHostname)
		if err != nil {
			log.Warn(err)
			return config, nil
		}

		config += fmt.Sprintf(
			tab+"StreamLocalBindUnlink yes\n"+
				tab+"RemoteForward %s:%s\n\n", remoteSocket, localSocket)
		err = RemoveSshEntry(profileId, workspaceId, projectName)
		if err != nil {
			log.Warn(err)
			return config, nil
		}
	} else {
		config += "\n"
	}

	return config, nil
}

func EnsureSshConfigEntryAdded(profileId, workspaceName, projectName string, gpgKey string) error {
	err := ensureSshFilesLinked()
	if err != nil {
		return err
	}

	sshDir := filepath.Join(sshHomeDir, ".ssh")
	configPath := filepath.Join(sshDir, "daytona_config")

	knownHostsFile := getKnownHostsFile()

	// Read existing content from the file
	existingContent, err := os.ReadFile(configPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Generate SSH config entry without GPG forwarding
	err = appendSshConfigEntry(configPath, profileId, workspaceName, projectName, knownHostsFile, false, existingContent)
	if err != nil {
		return err
	}

	if gpgKey != "" {
		// Generate SSH config entry with GPG forwarding and override previous config
		err = appendSshConfigEntry(configPath, profileId, workspaceName, projectName, knownHostsFile, true, existingContent)
		if err != nil {
			return err
		}
		projectHostname := GetProjectHostname(profileId, workspaceName, projectName)
		err = ExportGPGKey(gpgKey, projectHostname)
		if err != nil {
			return err
		}
	}

	return nil
}

func getKnownHostsFile() string {
	if runtime.GOOS == "windows" {
		return "NUL"
	}
	return "/dev/null"
}

func appendSshConfigEntry(configPath, profileId, workspaceId, projectName, knownHostsFile string, gpgForward bool, existingContent []byte) error {
	data, err := generateSshConfigEntry(profileId, workspaceId, projectName, knownHostsFile, gpgForward)
	if err != nil {
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
	return err
}

func getLocalGPGSocket() (string, error) {
	// Check if gpg is installed
	if _, err := exec.LookPath("gpg"); err != nil {
		return "", fmt.Errorf("gpg is not installed: %v", err)
	}

	// Attempt to get the local GPG socket
	cmd := exec.Command("gpgconf", "--list-dir", "agent-extra-socket")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get local GPG socket: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func getRemoteGPGSocket(projectHostname string) (string, error) {
	cmd := exec.Command("ssh", projectHostname, "gpgconf --list-dir agent-socket")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get remote GPG socket: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func ExportGPGKey(keyID, projectHostname string) error {
	exportCmd := exec.Command("gpg", "--export", keyID)
	var output bytes.Buffer
	exportCmd.Stdout = &output

	if err := exportCmd.Run(); err != nil {
		return err
	}

	importCmd := exec.Command("ssh", projectHostname, "gpg --import")
	importCmd.Stdin = &output

	return importCmd.Run()
}

func readSshConfig(configPath string) (string, error) {
	content, err := os.ReadFile(configPath)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	return string(content), nil
}

func writeSshConfig(configPath, newContent string) error {
	newContent = strings.Trim(newContent, "\n")

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

func RemoveSshEntry(profileId, workspaceId, projectName string) error {
	hostEntry := fmt.Sprintf("Host %s-%s-%s", profileId, workspaceId, projectName)

	sshDir := filepath.Join(os.Getenv("HOME"), ".ssh")
	configPath := filepath.Join(sshDir, "daytona_config")

	existingContent, err := readSshConfig(configPath)
	if err != nil {
		return err
	}

	// Define the regex pattern to match the specific Host entry
	regex := regexp.MustCompile(fmt.Sprintf(`%s\s*\n(?:\t.*\n?)*`, regexp.QuoteMeta(hostEntry)))

	// Replace the matched entry with an empty string
	newContent := regex.ReplaceAllString(existingContent, "")

	// Write the updated content back to the config file
	err = writeSshConfig(configPath, newContent)
	if err != nil {
		return err
	}

	return nil
}

// RemoveWorkspaceSshEntries removes all SSH entries for a given profileId and workspaceId
func RemoveWorkspaceSshEntries(profileId, workspaceId string) error {
	sshDir := filepath.Join(os.Getenv("HOME"), ".ssh")
	configPath := filepath.Join(sshDir, "daytona_config")

	// Read existing content from the SSH config file
	existingContent, err := readSshConfig(configPath)
	if err != nil {
		return err
	}

	// Define the regex pattern to match Host entries for the given profileId and workspaceId
	regex := regexp.MustCompile(fmt.Sprintf(`Host %s-%s-\w+\s*\n(?:\t.*\n?)*`, profileId, workspaceId))

	// Replace matched entries with an empty string
	newContent := regex.ReplaceAllString(existingContent, "")

	// Write the updated content back to the config file
	err = writeSshConfig(configPath, newContent)
	if err != nil {
		return err
	}

	return nil
}

func GetProjectHostname(profileId, workspaceId, projectName string) string {
	return fmt.Sprintf("%s-%s-%s", profileId, workspaceId, projectName)
}

func init() {
	if runtime.GOOS == "windows" {
		sshHomeDir = os.Getenv("USERPROFILE")
	} else {
		sshHomeDir = os.Getenv("HOME")
	}
}
