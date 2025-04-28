// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package agents

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

func InitClaude(homeDir string) (string, string, error) {
	var agentConfigFilePath string
	var mcpLogFilePath string

	switch runtime.GOOS {
	case "darwin":
		agentConfigFilePath = filepath.Join(homeDir, "Library", "Application Support", "Claude", "claude_desktop_config.json")
		mcpLogFilePath = filepath.Join(homeDir, "Library", "Logs", "Claude", mcpLogFileName)

	case "windows":
		// Resolve %APPDATA% environment variable
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", "", errors.New("could not resolve APPDATA environment variable")
		}

		agentConfigFilePath = filepath.Join(appData, "Claude", "claude_desktop_config.json")
		mcpLogFilePath = filepath.Join(appData, "Claude", "Logs", mcpLogFileName)

	case "linux":
		agentConfigFilePath = filepath.Join(homeDir, ".config", "Claude", "claude_desktop_config.json")
		mcpLogFilePath = filepath.Join("var", "log", "Claude", mcpLogFileName)
	default:
		return "", "", errors.New("operating system is not supported")
	}

	return agentConfigFilePath, mcpLogFilePath, nil
}
