//go:build !windows

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"io"
	"os"
	"path/filepath"

	"github.com/daytonaio/daytona/pkg/agent"
	"github.com/daytonaio/daytona/pkg/agent/config"
	"github.com/daytonaio/daytona/pkg/agent/ssh"
	"github.com/daytonaio/daytona/pkg/agent/tailscale"
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
	Run: func(cmd *cobra.Command, args []string) {
		if log.GetLevel() < log.InfoLevel {
			log.SetLevel(log.InfoLevel)
		}

		agentMode := config.ModeProject

		if hostModeFlag {
			agentMode = config.ModeHost
		}

		c, err := config.GetConfig(agentMode)
		if err != nil {
			log.Fatal(err)
		}
		c.ProjectDir = filepath.Join(os.Getenv("HOME"), c.ProjectName)

		gitLogWriter := io.MultiWriter(os.Stdout)
		var agentLogWriter io.Writer
		if c.LogFilePath != nil {
			logFile, err := os.OpenFile(*c.LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatal(err)
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

		tailscaleHostname := workspace.GetProjectHostname(c.WorkspaceId, c.ProjectName)
		if hostModeFlag {
			tailscaleHostname = c.WorkspaceId
		}

		tailscaleServer := &tailscale.Server{
			Hostname: tailscaleHostname,
			Server:   c.Server,
		}

		agent := agent.Agent{
			Config:                 c,
			Git:                    git,
			Ssh:                    sshServer,
			Tailscale:              tailscaleServer,
			LogWriter:              agentLogWriter,
			PostCreateLockFilePath: filepath.Join(os.Getenv("HOME"), ".daytona_post_create.lock"),
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
