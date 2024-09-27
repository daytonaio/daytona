// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspacemode

import (
	"os"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	cmd "github.com/daytonaio/daytona/pkg/cmd"
	. "github.com/daytonaio/daytona/pkg/cmd/agent"

	"github.com/spf13/cobra"
)

var workspaceId = ""
var projectName = ""

var workspaceModeRootCmd = &cobra.Command{
	Use:               "daytona",
	Short:             "Use the Daytona CLI to manage your workspace",
	Long:              "Use the Daytona CLI to manage your workspace",
	DisableAutoGenTag: true,
	SilenceUsage:      true,
	SilenceErrors:     true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func Execute() error {
	cmd.SetupRootCommand(workspaceModeRootCmd)
	workspaceModeRootCmd.AddGroup(&cobra.Group{ID: util.WORKSPACE_GROUP, Title: "Project & Workspace"})
	workspaceModeRootCmd.AddCommand(gitCredCmd)
	workspaceModeRootCmd.AddCommand(AgentCmd)
	workspaceModeRootCmd.AddCommand(startCmd)
	workspaceModeRootCmd.AddCommand(stopCmd)
	workspaceModeRootCmd.AddCommand(restartCmd)
	workspaceModeRootCmd.AddCommand(infoCmd)
	workspaceModeRootCmd.AddCommand(portForwardCmd)
	workspaceModeRootCmd.AddCommand(exposeCmd)

	clientId := config.GetClientId()
	telemetryEnabled := config.TelemetryEnabled()
	startTime := time.Now()

	telemetryService, command, flags, err := cmd.PreRun(workspaceModeRootCmd, os.Args[1:], telemetryEnabled, clientId, startTime)
	if err != nil {
		return err
	}

	err = workspaceModeRootCmd.Execute()

	endTime := time.Now()
	cmd.PostRun(command, err, telemetryService, clientId, startTime, endTime, flags)

	return err
}

func init() {
	if workspaceIdEnv := os.Getenv("DAYTONA_WS_ID"); workspaceIdEnv != "" {
		workspaceId = workspaceIdEnv
	}
	if projectNameEnv := os.Getenv("DAYTONA_WS_PROJECT_NAME"); projectNameEnv != "" {
		projectName = projectNameEnv
	}
}
