//go:build !windows

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"io"
	"os"
	"path"

	"github.com/daytonaio/daytona/pkg/agent"
	"github.com/daytonaio/daytona/pkg/agent/config"
	"github.com/daytonaio/daytona/pkg/agent/git"
	"github.com/daytonaio/daytona/pkg/agent/ssh"
	"github.com/daytonaio/daytona/pkg/agent/tailscale"
	"github.com/daytonaio/daytona/pkg/workspace"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var hostModeFlag bool

var AgentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Start the agent process",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if log.GetLevel() < log.InfoLevel {
			log.SetLevel(log.InfoLevel)
		}

		agentMode := config.ModeProject

		if hostModeFlag {
			agentMode = config.ModeHost
		}

		config, err := config.GetConfig(agentMode)
		if err != nil {
			log.Fatal(err)
		}
		config.ProjectDir = path.Join(os.Getenv("HOME"), config.ProjectName)

		gitLogWriter := io.MultiWriter(os.Stdout)
		var agentLogWriter io.Writer
		if config.LogFilePath != nil {
			logFile, err := os.OpenFile(*config.LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatal(err)
			}
			defer logFile.Close()
			gitLogWriter = io.MultiWriter(os.Stdout, logFile)
			agentLogWriter = logFile
		}

		git := &git.Service{
			ProjectDir:        config.ProjectDir,
			GitConfigFileName: path.Join(os.Getenv("HOME"), ".gitconfig"),
			LogWriter:         gitLogWriter,
		}

		sshServer := &ssh.Server{
			ProjectDir:        config.ProjectDir,
			DefaultProjectDir: os.Getenv("HOME"),
		}

		tailscaleHostname := workspace.GetProjectHostname(config.WorkspaceId, config.ProjectName)
		if hostModeFlag {
			tailscaleHostname = config.WorkspaceId
		}

		tailscaleServer := &tailscale.Server{
			Hostname:  tailscaleHostname,
			ServerUrl: config.Server.Url,
		}

		agent := agent.Agent{
			Config:    config,
			Git:       git,
			Ssh:       sshServer,
			Tailscale: tailscaleServer,
			LogWriter: agentLogWriter,
		}

		err = agent.Start()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	AgentCmd.Flags().BoolVar(&hostModeFlag, "host", false, "Run the agent in host mode")
	AgentCmd.AddCommand(logsCmd)
}
