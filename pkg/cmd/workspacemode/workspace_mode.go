// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspacemode

import (
	"fmt"
	"os"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal"
	"github.com/daytonaio/daytona/internal/util"
	cmd "github.com/daytonaio/daytona/pkg/cmd"
	. "github.com/daytonaio/daytona/pkg/cmd/agent"
	"github.com/daytonaio/daytona/pkg/posthogservice"
	"github.com/daytonaio/daytona/pkg/telemetry"
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

	var telemetryService telemetry.TelemetryService
	clientId := config.GetClientId()
	telemetryEnabled := config.TelemetryEnabled()

	if telemetryEnabled {
		telemetryService = posthogservice.NewTelemetryService(posthogservice.PosthogServiceConfig{
			ApiKey:   internal.PosthogApiKey,
			Endpoint: internal.PosthogEndpoint,
		})
	}

	command, err := cmd.ValidateCommands(workspaceModeRootCmd, os.Args[1:])
	if err != nil {
		fmt.Printf("Error: %v\n\n", err)
		helpErr := command.Help()
		if telemetryEnabled {
			props := cmd.GetCmdTelemetryData(command)
			props["command"] = os.Args[1]
			props["called_as"] = os.Args[1]
			err := telemetryService.TrackCliEvent(telemetry.CliEventInvalidCmd, clientId, props)
			if err != nil {
				log.Error(err)
			}
			telemetryService.Close()
		}

		return helpErr
	}

	if telemetryEnabled {
		err := telemetryService.TrackCliEvent(telemetry.CliEventCmdStart, clientId, cmd.GetCmdTelemetryData(command))
		if err != nil {
			log.Error(err)
		}
	}

	startTime := time.Now()

	err = workspaceModeRootCmd.Execute()

	endTime := time.Now()
	if telemetryService != nil {
		execTime := endTime.Sub(startTime)
		props := cmd.GetCmdTelemetryData(command)
		props["exec time (µs)"] = execTime.Microseconds()
		if err != nil {
			props["error"] = err.Error()
		}

		err := telemetryService.TrackCliEvent(telemetry.CliEventCmdEnd, clientId, props)
		if err != nil {
			log.Error(err)
		}
		telemetryService.Close()
	}

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
