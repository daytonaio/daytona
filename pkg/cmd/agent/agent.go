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
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/docker"
	"github.com/daytonaio/daytona/pkg/git"
	"github.com/daytonaio/daytona/pkg/models"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var targetModeFlag bool

var AgentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Start the agent process",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		setLogLevel()

		agentMode := agent_config.ModeWorkspace

		if targetModeFlag {
			agentMode = agent_config.ModeTarget
		}

		c, err := agent_config.GetConfig(agentMode)
		if err != nil {
			return err
		}

		telemetryEnabled := os.Getenv("DAYTONA_TELEMETRY_ENABLED") == "true"

		var ws *models.Workspace
		c.WorkspaceDir = os.Getenv("HOME")

		if agentMode == agent_config.ModeWorkspace {
			ws, err = getWorkspace(c, telemetryEnabled)
			if err != nil {
				return err
			}
			c.WorkspaceDir = filepath.Join(os.Getenv("HOME"), ws.WorkspaceFolderName())
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
		dockerCredHelperLogWriter := io.MultiWriter(os.Stdout)
		var agentLogWriter io.Writer
		if c.LogFilePath != nil {
			logFile, err := os.OpenFile(*c.LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			defer logFile.Close()
			gitLogWriter = io.MultiWriter(os.Stdout, logFile)
			dockerCredHelperLogWriter = io.MultiWriter(os.Stdout, logFile)
			agentLogWriter = logFile
		}

		git := &git.Service{
			WorkspaceDir:      c.WorkspaceDir,
			GitConfigFileName: filepath.Join(os.Getenv("HOME"), ".gitconfig"),
			LogWriter:         gitLogWriter,
		}

		dockerCredHelper := &docker.DockerCredHelper{
			DockerConfigFileName: filepath.Join(os.Getenv("HOME"), ".docker", "config.json"),
			LogWriter:            dockerCredHelperLogWriter,
			HomeDir:              os.Getenv("HOME"),
		}

		sshServer := &ssh.Server{
			WorkspaceDir:        c.WorkspaceDir,
			DefaultWorkspaceDir: os.Getenv("HOME"),
		}

		tailscaleHostname := common.GetTailscaleHostname(c.TargetId)
		if agentMode == agent_config.ModeWorkspace {
			tailscaleHostname = common.GetTailscaleHostname(ws.Id)
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
			DockerCredHelper: dockerCredHelper,
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
	AgentCmd.Flags().BoolVar(&targetModeFlag, "target", false, "Run the agent in target mode")
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

func getWorkspace(c *agent_config.Config, telemetryEnabled bool) (*models.Workspace, error) {
	ctx := context.Background()

	apiClient, err := apiclient_util.GetAgentApiClient(c.Server.ApiUrl, c.Server.ApiKey, c.ClientId, telemetryEnabled)
	if err != nil {
		return nil, err
	}

	workspace, res, err := apiClient.WorkspaceAPI.FindWorkspace(ctx, c.WorkspaceId).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	return conversion.Convert[apiclient.WorkspaceDTO, models.Workspace](workspace)
}
