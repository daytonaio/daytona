//go:build !windows

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"os"
	"path"

	"github.com/daytonaio/daytona/pkg/agent"
	"github.com/daytonaio/daytona/pkg/agent/config"
	"github.com/daytonaio/daytona/pkg/agent/git"
	"github.com/daytonaio/daytona/pkg/agent/ssh"
	"github.com/daytonaio/daytona/pkg/agent/tailscale"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var AgentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Start the agent process",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		git := &git.Service{
			ProjectDir:        config.ProjectDir,
			GitConfigFileName: path.Join(os.Getenv("HOME"), ".gitconfig"),
		}

		sshServer := &ssh.Server{
			ProjectDir:        config.ProjectDir,
			DefaultProjectDir: os.Getenv("HOME"),
		}

		tailscaleServer := &tailscale.Server{
			WorkspaceId: config.WorkspaceId,
			ProjectName: config.ProjectName,
			ServerUrl:   config.Server.Url,
		}

		agent := agent.Agent{
			Config:    config,
			Git:       git,
			Ssh:       sshServer,
			Tailscale: tailscaleServer,
		}

		err = agent.Start()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {

}
