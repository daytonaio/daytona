//go:build !windows

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"context"
	"io"
	"os"
	"path/filepath"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/conversion"
	"github.com/daytonaio/daytona/pkg/agent"
	"github.com/daytonaio/daytona/pkg/agent/config"
	"github.com/daytonaio/daytona/pkg/agent/ssh"
	"github.com/daytonaio/daytona/pkg/agent/tailscale"
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

		agentMode := config.ModeWorkspace

		if targetModeFlag {
			agentMode = config.ModeTarget
		}

		c, err := config.GetConfig(agentMode)
		if err != nil {
			return err
		}

		telemetryEnabled := os.Getenv("DAYTONA_TELEMETRY_ENABLED") == "true"

		var ws *models.Workspace

		if agentMode == config.ModeWorkspace {
			ws, err = getWorkspace(c, telemetryEnabled)
			if err != nil {
				return err
			}

			c.WorkspaceDir = filepath.Join(os.Getenv("HOME"), ws.WorkspaceFolderName())
		}

		if workspaceDir := os.Getenv("DAYTONA_WORKSPACE_DIR"); workspaceDir != "" {
			c.WorkspaceDir = workspaceDir
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
		if agentMode == config.ModeWorkspace {
			tailscaleHostname = common.GetTailscaleHostname(ws.Id)
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

func getWorkspace(c *config.Config, telemetryEnabled bool) (*models.Workspace, error) {
	ctx := context.Background()

	apiClient, err := apiclient_util.GetAgentApiClient(c.Server.ApiUrl, c.Server.ApiKey, c.ClientId, telemetryEnabled)
	if err != nil {
		return nil, err
	}

	workspace, res, err := apiClient.WorkspaceAPI.GetWorkspace(ctx, c.WorkspaceId).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	return conversion.Convert[apiclient.WorkspaceDTO, models.Workspace](workspace)
}
