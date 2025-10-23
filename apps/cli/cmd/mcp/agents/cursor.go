// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package agents

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/daytonaio/daytona/cli/cmd/mcp/common"
)

func InitCursor(homeDir, mcpServerName string) (string, string, error) {
	var agentConfigFilePath string
	var mcpLogFilePath string

	mcpLogFileName := fmt.Sprintf(common.MCP_LOG_FILE_NAME_FORMAT, mcpServerName)

	switch runtime.GOOS {
	case "darwin":
		agentConfigFilePath = filepath.Join(homeDir, ".cursor", "mcp.json")
		mcpLogFilePath = filepath.Join(homeDir, "Library", "Logs", "Cursor", mcpLogFileName)

	case "windows":
		// Resolve %APPDATA% environment variable
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", "", errors.New("could not resolve APPDATA environment variable")
		}

		agentConfigFilePath = filepath.Join(appData, ".cursor", "mcp.json")
		mcpLogFilePath = filepath.Join(appData, "Cursor", "Logs", mcpLogFileName)

	case "linux":
		agentConfigFilePath = filepath.Join(homeDir, ".cursor", "mcp.json")
		mcpLogFilePath = filepath.Join("var", "log", "Cursor", mcpLogFileName)
	default:
		return "", "", errors.New("operating system is not supported")
	}

	return agentConfigFilePath, mcpLogFilePath, nil
}
