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

var SshHomeDir string

func ensureSshFilesLinked() error {
	// Make sure ~/.ssh/config file exists if not create it
	sshDir := filepath.Join(SshHomeDir, ".ssh")
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

		contentStr := string(content)

		if strings.HasPrefix(contentStr, "Include daytona_config") {
			return nil
		}

		newContent := strings.ReplaceAll(contentStr, "Include daytona_config\n\n", "")
		newContent = strings.ReplaceAll(newContent, "Include daytona_config\n", "")
		newContent = strings.ReplaceAll(newContent, "Include daytona_config", "")
		newContent = "Include daytona_config\n\n" + newContent
		err = os.WriteFile(configFile, []byte(newContent), 0600)
		if err != nil {
			return err
		}
	}

	return nil
}

func UnlinkSshFiles() error {
	sshDirPath := filepath.Join(SshHomeDir, ".ssh")
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

func generateSshConfigEntry(profileId, workspaceName, knownHostsPath string, gpgForward bool) (string, error) {
	daytonaPath, err := os.Executable()
	if err != nil {
		return "", err
	}

	tab := "\t"
	workspaceHostname := GetWorkspaceHostname(profileId, workspaceName)

	config := fmt.Sprintf("Host %s\n"+
		tab+"User daytona\n"+
		tab+"StrictHostKeyChecking no\n"+
		tab+"UserKnownHostsFile %s\n"+
		tab+"ProxyCommand \"%s\" ssh-proxy %s %s\n"+
		tab+"ForwardAgent yes\n", workspaceHostname, knownHostsPath, daytonaPath, profileId, workspaceName)

	if gpgForward {
		localSocket, err := getLocalGPGSocket()
		if err != nil {
			log.Trace(err)
			return config, nil
		}

		remoteSocket, err := getRemoteGPGSocket(workspaceHostname)
		if err != nil {
			log.Trace(err)
			return config, nil
		}

		config += fmt.Sprintf(
			tab+"StreamLocalBindUnlink yes\n"+
				tab+"RemoteForward %s:%s\n\n", remoteSocket, localSocket)
	} else {
		config += "\n"
	}

	return config, nil
}

func EnsureSshConfigEntryAdded(profileId, workspaceId string, gpgKey string) error {
	err := ensureSshFilesLinked()
	if err != nil {
		return err
	}

	sshDir := filepath.Join(SshHomeDir, ".ssh")
	configPath := filepath.Join(sshDir, "daytona_config")

	knownHostsFile := getKnownHostsFile()

	// Read existing content from the file
	existingContent, err := ReadSshConfig(configPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	var configGenerated bool
	regexWithoutGPG := regexp.MustCompile(fmt.Sprintf(`(?m)^Host %s-%s\s*\n(?:\s+[^\n]*\n?)*`, profileId, workspaceId))
	regexWithGPG := regexp.MustCompile(fmt.Sprintf(`(?m)^Host %s-%s\s*\n(?:\s+[^\n]*\n?)*StreamLocalBindUnlink\s+yes\s*\n(?:\s+[^\n]*\n?)*RemoteForward\s+[^\s]+\s+[^\s]+\s*\n`, profileId, workspaceId))
	if !regexWithoutGPG.MatchString(existingContent) {
		newContent, err := appendSshConfigEntry(configPath, profileId, workspaceId, knownHostsFile, false, existingContent)
		if err != nil {
			return err
		}
		existingContent = newContent
		configGenerated = true
	}

	if gpgKey != "" && !regexWithGPG.MatchString(existingContent) {
		_, err := appendSshConfigEntry(configPath, profileId, workspaceId, knownHostsFile, true, existingContent)
		if err != nil {
			return err
		}

		workspaceHostname := GetWorkspaceHostname(profileId, workspaceId)
		err = ExportGPGKey(gpgKey, workspaceHostname)
		if err != nil {
			return err
		}

		configGenerated = true
	}

	if !configGenerated {
		updatedContent, err := regenerateProxyCommand(existingContent, profileId, workspaceId)
		if err != nil {
			return err
		}
		err = UpdateWorkspaceSshEntry(profileId, workspaceId, updatedContent)
		if err != nil {
			return err
		}
	}

	return nil
}

func regenerateProxyCommand(existingContent, profileId, workspaceId string) (string, error) {
	daytonaPath, err := os.Executable()
	if err != nil {
		return "", err
	}

	hostLine := fmt.Sprintf("Host %s", GetWorkspaceHostname(profileId, workspaceId))
	regex := regexp.MustCompile(fmt.Sprintf(`%s\s*\n(?:\t.*\n?)*`, hostLine))
	matchedEntry := regex.FindString(existingContent)
	if matchedEntry == "" {
		return "", fmt.Errorf("no SSH entry found for workspace %s", workspaceId)
	}

	re := regexp.MustCompile(`(?m)^\s*ProxyCommand\s+.*$`)
	updatedContent := re.ReplaceAllString(matchedEntry, fmt.Sprintf("\tProxyCommand \"%s\" ssh-proxy %s %s", daytonaPath, profileId, workspaceId))

	return updatedContent, nil
}

func getKnownHostsFile() string {
	if runtime.GOOS == "windows" {
		return "NUL"
	}
	return "/dev/null"
}

func appendSshConfigEntry(configPath, profileId, workspaceId, knownHostsFile string, gpgForward bool, existingContent string) (string, error) {
	data, err := generateSshConfigEntry(profileId, workspaceId, knownHostsFile, gpgForward)
	if err != nil {
		return "", err
	}

	if strings.Contains(existingContent, data) {
		// Entry already exists in the file
		return existingContent, nil
	}

	// We want to remove the config entry gpg counterpart
	configCounterpart, err := generateSshConfigEntry(profileId, workspaceId, knownHostsFile, !gpgForward)
	if err != nil {
		return "", err
	}
	updatedContent := strings.ReplaceAll(existingContent, configCounterpart, "")
	updatedContent = data + updatedContent

	// Open the file for writing
	file, err := os.Create(configPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = file.WriteString(updatedContent)
	return updatedContent, err
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

func getRemoteGPGSocket(workspaceHostname string) (string, error) {
	cmd := exec.Command("ssh", workspaceHostname, "gpgconf --list-dir agent-socket")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get remote GPG socket: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func ExportGPGKey(keyID, workspaceHostname string) error {
	exportCmd := exec.Command("gpg", "--export", keyID)
	var output bytes.Buffer
	exportCmd.Stdout = &output

	if err := exportCmd.Run(); err != nil {
		return err
	}

	importCmd := exec.Command("ssh", workspaceHostname, "gpg --import")
	importCmd.Stdin = &output

	return importCmd.Run()
}

func ReadSshConfig(configPath string) (string, error) {
	content, err := os.ReadFile(configPath)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	return string(content), nil
}

func writeSshConfig(configPath, newContent string) error {
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

// RemoveTargetSshEntries removes all SSH entries for a given profileId and workspaceId
func RemoveWorkspaceSshEntries(profileId, workspaceId string) error {
	sshDir := filepath.Join(SshHomeDir, ".ssh")
	configPath := filepath.Join(sshDir, "daytona_config")

	// Read existing content from the SSH config file
	existingContent, err := ReadSshConfig(configPath)
	if err != nil {
		return err
	}

	// Define the regex pattern to match Host entries for the given profileId and workspaceId
	regex := regexp.MustCompile(fmt.Sprintf(`Host %s-%s\n(?:\t.*\n?)*`, profileId, workspaceId))

	contentToDelete := regex.FindString(existingContent)
	if contentToDelete == "" {
		return nil
	}

	newContent := strings.ReplaceAll(existingContent, contentToDelete, "")
	newContent = strings.TrimSpace(newContent)

	// Write the updated content back to the config file
	return writeSshConfig(configPath, newContent)
}

func UpdateWorkspaceSshEntry(profileId, workspaceId, updatedContent string) error {
	sshDir := filepath.Join(SshHomeDir, ".ssh")
	configPath := filepath.Join(sshDir, "daytona_config")

	existingContent, err := ReadSshConfig(configPath)
	if err != nil {
		return err
	}

	hostLine := fmt.Sprintf("Host %s", GetWorkspaceHostname(profileId, workspaceId))
	regex := regexp.MustCompile(fmt.Sprintf(`%s\s*\n(?:\t.*\n?)*`, hostLine))
	oldContent := regex.FindString(existingContent)
	if oldContent == "" {
		return fmt.Errorf("no SSH entry found for workspace %s", workspaceId)
	}
	existingContent = strings.ReplaceAll(existingContent, oldContent, updatedContent)

	err = writeSshConfig(configPath, existingContent)
	if err != nil {
		return err
	}

	return nil
}

func GetWorkspaceHostname(profileId, workspaceName string) string {
	return fmt.Sprintf("%s-%s", profileId, workspaceName)
}

func init() {
	if runtime.GOOS == "windows" {
		SshHomeDir = os.Getenv("USERPROFILE")
	} else {
		SshHomeDir = os.Getenv("HOME")
	}
}
