// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package mcp

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/daytonaio/daytona/cli/cmd/mcp/agents"
	"github.com/spf13/cobra"
)

var InitCmd = &cobra.Command{
	Use:   "init [AGENT_NAME]",
	Short: "Initialize Daytona MCP Server with an agent (currently supported: claude, windsurf, cursor)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("agent name is required")
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		var agentConfigFilePath, mcpLogFilePath string

		switch args[0] {
		case "claude":
			agentConfigFilePath, mcpLogFilePath, err = agents.InitClaude(homeDir)
			if err != nil {
				return err
			}
		case "cursor":
			agentConfigFilePath, mcpLogFilePath, err = agents.InitCursor(homeDir)
			if err != nil {
				return err
			}
		case "windsurf":
			agentConfigFilePath, mcpLogFilePath, err = agents.InitWindsurf(homeDir)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("agent name %s is not supported", args[0])
		}

		return injectConfig(agentConfigFilePath, mcpLogFilePath)
	},
}

func injectConfig(agentConfigFilePath, mcpLogFilePath string) error {
	daytonaMcpConfig, err := getDayonaMcpConfig(mcpLogFilePath)
	if err != nil {
		return err
	}

	// Read existing model config or create new one
	var agentConfig map[string]interface{}
	if agentConfigData, err := os.ReadFile(agentConfigFilePath); err == nil {
		if err := json.Unmarshal(agentConfigData, &agentConfig); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	} else {
		agentConfig = make(map[string]interface{})
	}

	// Initialize or update mcpServers field
	mcpServers, ok := agentConfig["mcpServers"].(map[string]interface{})
	if !ok {
		mcpServers = make(map[string]interface{})
	}

	// Add or update daytona-mcp configuration
	mcpServers["daytona-mcp"] = daytonaMcpConfig
	agentConfig["mcpServers"] = mcpServers

	// Write back the updated config with indentation
	updatedJSON, err := json.MarshalIndent(agentConfig, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(agentConfigFilePath, updatedJSON, 0644)
}
