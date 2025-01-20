//go:build !windows

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/agent"
	agent_config "github.com/daytonaio/daytona/pkg/agent/config"
	"github.com/daytonaio/daytona/pkg/agent/ssh"
	"github.com/daytonaio/daytona/pkg/agent/tailscale"
	"github.com/daytonaio/daytona/pkg/agent/toolbox"
	"github.com/daytonaio/daytona/pkg/git"
	"github.com/daytonaio/daytona/pkg/workspace/project"
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

		agentMode := agent_config.ModeProject

		if hostModeFlag {
			agentMode = agent_config.ModeHost
		}

		c, err := agent_config.GetConfig(agentMode)
		if err != nil {
			return err
		}
		c.ProjectDir = filepath.Join(os.Getenv("HOME"), c.ProjectName)

		configDir, err := config.GetConfigDir()
		if err != nil {
			return err
		}

		if projectDir := os.Getenv("DAYTONA_PROJECT_DIR"); projectDir != "" {
			c.ProjectDir = projectDir
		}

		if _, err := os.Stat(c.ProjectDir); os.IsNotExist(err) {
			if err := os.MkdirAll(c.ProjectDir, 0755); err != nil {
				return fmt.Errorf("failed to create project directory: %w", err)
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
			ProjectDir:        c.ProjectDir,
			GitConfigFileName: filepath.Join(os.Getenv("HOME"), ".gitconfig"),
			LogWriter:         gitLogWriter,
		}

		sshServer := &ssh.Server{
			ProjectDir:        c.ProjectDir,
			DefaultProjectDir: os.Getenv("HOME"),
		}

		tailscaleHostname := project.GetProjectHostname(c.WorkspaceId, c.ProjectName)
		if hostModeFlag {
			tailscaleHostname = c.WorkspaceId
		}

		toolBoxServer := &toolbox.Server{
			ProjectDir: c.ProjectDir,
			ConfigDir:  configDir,
		}

		telemetryEnabled := os.Getenv("DAYTONA_TELEMETRY_ENABLED") == "true"

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
