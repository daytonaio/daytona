// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"regexp"
)

const completionScriptNameRoot = "daytona.completion_script."

var shellNames = []string{"bash", "zsh", "fish", "powershell"}

func DeleteAutocompletionData() error {
	for _, shellName := range shellNames {
		err := removeAutocompletionDataForShell(shellName)
		if err != nil {
			return err
		}
	}

	return nil
}

func removeAutocompletionDataForShell(shellName string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	completionScriptPath := homeDir
	runCommandFilePath := homeDir

	switch shellName {
	case "bash":
		completionScriptPath += "/." + completionScriptNameRoot + "bash"
		runCommandFilePath += "/.bashrc"
	case "zsh":
		completionScriptPath += "/." + completionScriptNameRoot + "zsh"
		runCommandFilePath += "/.zshrc"
	case "fish":
		completionScriptPath += "/." + completionScriptNameRoot + "fish"
		runCommandFilePath += "/.config/fish/config.fish"
	case "powershell":
		completionScriptPath += "/" + completionScriptNameRoot + "ps1"
		runCommandFilePath += "/Documents/WindowsPowerShell/Microsoft.PowerShell_profile.ps1"
	default:
		return nil
	}

	// Remove the line that sources the completion script if it exists
	err = removeLineFromFile(runCommandFilePath, completionScriptNameRoot)
	if err != nil {
		return err
	}

	// Remove the completion script if it exists
	_, err = os.Stat(completionScriptPath)
	if os.IsNotExist(err) {
		return nil
	}

	return os.Remove(completionScriptPath)
}

func removeLineFromFile(filePath string, lineText string) error {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	re := regexp.MustCompile("(?m:^.*" + regexp.QuoteMeta(lineText) + ".*$\n?)")
	newContent := re.ReplaceAllString(string(content), "")

	err = os.WriteFile(filePath, []byte(newContent), 0600)
	if err != nil {
		return err
	}

	return nil
}
