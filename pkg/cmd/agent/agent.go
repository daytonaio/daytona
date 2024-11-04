//go:build !windows

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/conversion"
	"github.com/daytonaio/daytona/pkg/agent"
	agent_config "github.com/daytonaio/daytona/pkg/agent/config"
	"github.com/daytonaio/daytona/pkg/agent/ssh"
	"github.com/daytonaio/daytona/pkg/agent/tailscale"
	"github.com/daytonaio/daytona/pkg/agent/toolbox"
	"github.com/daytonaio/daytona/pkg/git"
	"github.com/daytonaio/daytona/pkg/workspace"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var hostModeFlag bool

var AgentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Start the agent process",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		setLogLevel()

		agentMode := agent_config.ModeWorkspace

		if hostModeFlag {
			agentMode = agent_config.ModeHost
		}

		c, err := agent_config.GetConfig(agentMode)
		if err != nil {
			return err
		}

		telemetryEnabled := os.Getenv("DAYTONA_TELEMETRY_ENABLED") == "true"

		var ws *workspace.Workspace
		c.WorkspaceDir = os.Getenv("HOME")

		if agentMode == agent_config.ModeWorkspace {
			ws, err = getWorkspace(c, telemetryEnabled)
			if err != nil {
				return err
			}
			c.WorkspaceDir = filepath.Join(os.Getenv("HOME"), ws.Name)
		}

		configDir, err := config.GetConfigDir()
		if err != nil {
			return err
		}

		if workspaceDir := os.Getenv("DAYTONA_WORKSPACE_DIR"); workspaceDir != "" {
			c.WorkspaceDir = workspaceDir
		}

		if _, err := os.Stat(c.WorkspaceDir); os.IsNotExist(err) {
			if err := os.MkdirAll(c.WorkspaceDir, 0755); err != nil {
				return fmt.Errorf("failed to create workspace directory: %w", err)
			}
		}

		gitLogWriter := io.MultiWriter(os.Stdout)
		var agentLogWriter io.Writer
		if c.LogFilePath != nil {
			logFile, err := os.OpenFile(*c.LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			defer logFile.Close()
			gitLogWriter = io.MultiWriter(os.Stdout, logFile)
			agentLogWriter = logFile
		}

		git := &git.Service{
			WorkspaceDir:      c.WorkspaceDir,
			GitConfigFileName: filepath.Join(os.Getenv("HOME"), ".gitconfig"),
			LogWriter:         gitLogWriter,
		}

		sshServer := &ssh.Server{
			WorkspaceDir:        c.WorkspaceDir,
			DefaultWorkspaceDir: os.Getenv("HOME"),
		}

		tailscaleHostname := workspace.GetWorkspaceHostname(c.WorkspaceId)
		if hostModeFlag {
			tailscaleHostname = c.TargetId
		}

		toolBoxServer := &toolbox.Server{
			WorkspaceDir: c.WorkspaceDir,
			ConfigDir:    configDir,
		}

		tailscaleServer := &tailscale.Server{
			Hostname:         tailscaleHostname,
			Server:           c.Server,
			TelemetryEnabled: telemetryEnabled,
			ClientId:         c.ClientId,
		}

		agent := agent.Agent{
			Config:           c,
			Git:              git,
			Ssh:              sshServer,
			Toolbox:          toolBoxServer,
			Tailscale:        tailscaleServer,
			LogWriter:        agentLogWriter,
			TelemetryEnabled: telemetryEnabled,
			Workspace:        ws,
		}

		return agent.Start()
	},
}

func init() {
	AgentCmd.Flags().BoolVar(&hostModeFlag, "host", false, "Run the agent in host mode")
	AgentCmd.AddCommand(logsCmd)
}

func setLogLevel() {
	agentLogLevel := os.Getenv("AGENT_LOG_LEVEL")
	if agentLogLevel != "" {
		level, err := log.ParseLevel(agentLogLevel)
		if err != nil {
			log.Errorf("Invalid log level: %s, defaulting to info level", agentLogLevel)
			level = log.InfoLevel
		}
		log.SetLevel(level)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

func getWorkspace(c *agent_config.Config, telemetryEnabled bool) (*workspace.Workspace, error) {
	ctx := context.Background()

	apiClient, err := apiclient_util.GetAgentApiClient(c.Server.ApiUrl, c.Server.ApiKey, c.ClientId, telemetryEnabled)
	if err != nil {
		return nil, err
	}

	workspace, res, err := apiClient.WorkspaceAPI.GetWorkspace(ctx, c.WorkspaceId).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	return conversion.ToWorkspace(workspace), nil
}
