// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspacemode

import (
	"os"

	"github.com/daytonaio/daytona/internal/util"
	cmd "github.com/daytonaio/daytona/pkg/cmd"
	. "github.com/daytonaio/daytona/pkg/cmd/agent"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var workspaceId = ""
var projectName = ""

var workspaceModeRootCmd = &cobra.Command{
	Use:               "daytona",
	Short:             "Use the Daytona CLI to manage your workspace",
	Long:              "Use the Daytona CLI to manage your workspace",
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func Execute() {
	cmd.SetupRootCommand(workspaceModeRootCmd)
	workspaceModeRootCmd.AddGroup(&cobra.Group{ID: util.WORKSPACE_GROUP, Title: "Project & Workspace"})
	workspaceModeRootCmd.AddCommand(gitCredCmd)
	workspaceModeRootCmd.AddCommand(AgentCmd)
	workspaceModeRootCmd.AddCommand(startCmd)
	workspaceModeRootCmd.AddCommand(stopCmd)
	workspaceModeRootCmd.AddCommand(infoCmd)
	workspaceModeRootCmd.AddCommand(portForwardCmd)

	if err := workspaceModeRootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	if workspaceIdEnv := os.Getenv("DAYTONA_WS_ID"); workspaceIdEnv != "" {
		workspaceId = workspaceIdEnv
	}
	if projectNameEnv := os.Getenv("DAYTONA_WS_PROJECT_NAME"); projectNameEnv != "" {
		projectName = projectNameEnv
	}
}
