// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/daytonaio/daytona/cli/cmd/mcp/common"
	"github.com/spf13/cobra"
)

var ConfigCmd = &cobra.Command{
	Use:   "config [MCP_SERVER_NAME]",
	Short: "Outputs JSON configuration for Daytona MCP Server",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mcpServerName := ""
		if len(args) == 1 {
			mcpServerName = args[0]
		}

		mcpLogFileName := "daytona-mcp.log"
		if mcpServerName != "" {
			mcpLogFileName = fmt.Sprintf(common.MCP_LOG_FILE_NAME_FORMAT, mcpServerName)
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		var mcpLogFilePath string

		switch runtime.GOOS {
		case "darwin":
			mcpLogFilePath = homeDir + "/.daytona/" + mcpLogFileName
		case "windows":
			mcpLogFilePath = os.Getenv("APPDATA") + "\\.daytona\\" + mcpLogFileName
		case "linux":
			mcpLogFilePath = homeDir + "/.daytona/" + mcpLogFileName
		default:
			return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
		}

		daytonaMcpConfig, err := getDayonaMcpConfig(mcpLogFilePath, mcpServerName)
		if err != nil {
			return err
		}

		configServerName := "daytona-mcp"
		if mcpServerName != "" {
			configServerName = fmt.Sprintf("daytona-%s-mcp-server", mcpServerName)
		}

		mcpConfig := map[string]interface{}{
			configServerName: daytonaMcpConfig,
		}

		jsonBytes, err := json.MarshalIndent(mcpConfig, "", "  ")
		if err != nil {
			return err
		}

		fmt.Println(string(jsonBytes))

		return nil
	},
}

func getDayonaMcpConfig(mcpLogFilePath, mcpServerName string) (map[string]interface{}, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	args := []string{"mcp", "serve"}
	if mcpServerName != "" {
		args = append(args, mcpServerName)
	}

	// Create daytona-mcp config
	daytonaMcpConfig := map[string]interface{}{
		"command": "daytona",
		"args":    args,
		"env": map[string]string{
			"PATH": homeDir + ":/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:/opt/homebrew/bin",
			"HOME": homeDir,
		},
		"logFile": mcpLogFilePath,
	}

	if runtime.GOOS == "windows" {
		daytonaMcpConfig["env"].(map[string]string)["APPDATA"] = os.Getenv("APPDATA")
	}

	return daytonaMcpConfig, nil
}
