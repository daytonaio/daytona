// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agentmode

import (
	"os"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	cmd "github.com/daytonaio/daytona/pkg/cmd"
	. "github.com/daytonaio/daytona/pkg/cmd/agent"

	"github.com/spf13/cobra"
)

var targetId = ""
var workspaceId = ""

var agentModeRootCmd = &cobra.Command{
	Use:               "daytona",
	Short:             "Daytona is a Dev Environment Manager",
	Long:              "Daytona is a Dev Environment Manager",
	DisableAutoGenTag: true,
	SilenceUsage:      true,
	SilenceErrors:     true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func Execute() error {
	cmd.SetupRootCommand(agentModeRootCmd)
	agentModeRootCmd.AddGroup(&cobra.Group{ID: util.TARGET_GROUP, Title: "Workspace & Target"})
	agentModeRootCmd.AddCommand(gitCredCmd)
	agentModeRootCmd.AddCommand(dockerCredCmd)
	agentModeRootCmd.AddCommand(AgentCmd)
	agentModeRootCmd.AddCommand(infoCmd)
	agentModeRootCmd.AddCommand(portForwardCmd)
	agentModeRootCmd.AddCommand(exposeCmd)
	agentModeRootCmd.AddCommand(logsCmd)

	clientId := config.GetClientId()
	telemetryEnabled := config.TelemetryEnabled()
	startTime := time.Now()

	telemetryService, command, flags, isComplete, err := cmd.PreRun(agentModeRootCmd, os.Args[1:], telemetryEnabled, clientId, startTime)
	if err != nil {
		return err
	}

	err = agentModeRootCmd.Execute()

	endTime := time.Now()
	if !isComplete {
		cmd.PostRun(command, err, telemetryService, clientId, startTime, endTime, flags)
	}

	return err
}

func init() {
	if targetIdEnv := os.Getenv("DAYTONA_TARGET_ID"); targetIdEnv != "" {
		targetId = targetIdEnv
	}
	if workspaceIdEnv := os.Getenv("DAYTONA_WORKSPACE_ID"); workspaceIdEnv != "" {
		workspaceId = workspaceIdEnv
	}
}

func isWorkspaceAgentMode() bool {
	return workspaceId != ""
}
