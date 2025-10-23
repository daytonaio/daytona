// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"

	"github.com/daytonaio/daytona/cli/cmd/mcp/agents"
	"github.com/daytonaio/daytona/cli/cmd/mcp/common"
	"github.com/spf13/cobra"
)

var InitCmd = &cobra.Command{
	Use:   "init [MCP_SERVER_NAME] [AGENT_NAME]",
	Short: "Initialize any Daytona MCP Server with an agent. Currently available Daytona MCP Servers: <empty> for daytona code execution MCP, 'sandbox' for Sandbox actions MCP, 'fs' for Filesystem operations MCP, 'git' for Git operations MCP; currently supported agents: 'claude', 'windsurf', 'cursor'",
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("mcp server name and agent name are required")
		}

		var mcpServerName, agentName string

		if len(args) == 1 {
			if !slices.Contains(common.SupportedAgents, args[0]) || !slices.Contains(common.SupportedDaytonaMCPServers, args[0]) {
				return fmt.Errorf("agent name %s is not supported", args[0])
			}

			agentName = args[0]
			mcpServerName = ""
		}

		if len(args) == 2 {
			if !slices.Contains(common.SupportedDaytonaMCPServers, args[0]) || !slices.Contains(common.SupportedAgents, args[1]) {
				return fmt.Errorf("mcp server name %s or agent name %s is not supported", args[0], args[1])
			}

			mcpServerName = args[0]
			agentName = args[1]
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		mcpLogFileName := "daytona-mcp.log"
		if mcpServerName != "" {
			mcpLogFileName = fmt.Sprintf(common.MCP_LOG_FILE_NAME_FORMAT, mcpServerName)
		}

		var agentConfigFilePath, mcpLogFilePath string

		switch agentName {
		case "claude":
			agentConfigFilePath, mcpLogFilePath, err = agents.InitClaude(homeDir, mcpLogFileName)
			if err != nil {
				return err
			}
		case "cursor":
			agentConfigFilePath, mcpLogFilePath, err = agents.InitCursor(homeDir, mcpLogFileName)
			if err != nil {
				return err
			}
		case "windsurf":
			agentConfigFilePath, mcpLogFilePath, err = agents.InitWindsurf(homeDir, mcpLogFileName)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("agent name %s is not supported", agentName)
		}

		return injectConfig(agentConfigFilePath, mcpLogFilePath, mcpServerName)
	},
}

func injectConfig(agentConfigFilePath, mcpLogFilePath, mcpServerName string) error {
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

	configServerName := "daytona-mcp"
	if mcpServerName != "" {
		configServerName = fmt.Sprintf("daytona-%s-mcp-server", mcpServerName)
	}

	// Add or update daytona-mcp configuration
	mcpServers[configServerName] = daytonaMcpConfig
	agentConfig["mcpServers"] = mcpServers

	// Write back the updated config with indentation
	updatedJSON, err := json.MarshalIndent(agentConfig, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(agentConfigFilePath, updatedJSON, 0644)
}
