// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"
	"strings"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal"
	. "github.com/daytonaio/daytona/internal/util"
	. "github.com/daytonaio/daytona/pkg/cmd/apikey"
	. "github.com/daytonaio/daytona/pkg/cmd/containerregistry"
	. "github.com/daytonaio/daytona/pkg/cmd/gitprovider"
	"github.com/daytonaio/daytona/pkg/cmd/output"
	. "github.com/daytonaio/daytona/pkg/cmd/ports"
	. "github.com/daytonaio/daytona/pkg/cmd/profile"
	. "github.com/daytonaio/daytona/pkg/cmd/profiledata/env"
	. "github.com/daytonaio/daytona/pkg/cmd/provider"
	. "github.com/daytonaio/daytona/pkg/cmd/server"
	. "github.com/daytonaio/daytona/pkg/cmd/target"
	. "github.com/daytonaio/daytona/pkg/cmd/telemetry"
	. "github.com/daytonaio/daytona/pkg/cmd/workspace"
	"github.com/daytonaio/daytona/pkg/posthogservice"
	"github.com/daytonaio/daytona/pkg/telemetry"
	view "github.com/daytonaio/daytona/pkg/views/initial"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:               "daytona",
	Short:             "Daytona is a Dev Environment Manager",
	Long:              "Daytona is a Dev Environment Manager",
	DisableAutoGenTag: true,
	Run:               RunInitialScreenFlow,
}

var originalStdout *os.File

func Execute() {
	rootCmd.AddGroup(&cobra.Group{ID: WORKSPACE_GROUP, Title: "Workspaces & Projects"})
	rootCmd.AddGroup(&cobra.Group{ID: SERVER_GROUP, Title: "Server"})
	rootCmd.AddGroup(&cobra.Group{ID: PROFILE_GROUP, Title: "Profile"})

	rootCmd.AddCommand(CodeCmd)
	rootCmd.AddCommand(SshCmd)
	rootCmd.AddCommand(SshProxyCmd)
	rootCmd.AddCommand(CreateCmd)
	rootCmd.AddCommand(DeleteCmd)
	rootCmd.AddCommand(ServeCmd)
	rootCmd.AddCommand(ServerCmd)
	rootCmd.AddCommand(ApiKeyCmd)
	rootCmd.AddCommand(ContainerRegistryCmd)
	rootCmd.AddCommand(ProviderCmd)
	rootCmd.AddCommand(TargetCmd)
	rootCmd.AddCommand(ideCmd)
	rootCmd.AddCommand(ProfileCmd)
	rootCmd.AddCommand(ProfileUseCmd)
	rootCmd.AddCommand(whoamiCmd)
	rootCmd.AddCommand(purgeCmd)
	rootCmd.AddCommand(GitProviderCmd)
	rootCmd.AddCommand(StartCmd)
	rootCmd.AddCommand(StopCmd)
	rootCmd.AddCommand(InfoCmd)
	rootCmd.AddCommand(PortForwardCmd)
	rootCmd.AddCommand(EnvCmd)
	rootCmd.AddCommand(TelemetryCmd)

	SetupRootCommand(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func SetupRootCommand(cmd *cobra.Command) {
	// Common commands
	cmd.AddCommand(AutoCompleteCmd)
	cmd.AddCommand(versionCmd)
	cmd.AddCommand(ListCmd)
	cmd.AddCommand(generateDocsCmd)
	cmd.AddCommand(DocsCmd)

	cmd.CompletionOptions.HiddenDefaultCmd = true
	cmd.PersistentFlags().BoolP("help", "", false, "help for daytona")
	cmd.PersistentFlags().StringVarP(&output.FormatFlag, "output", "o", output.FormatFlag, `Output format. Must be one of (yaml, json)`)

	var telemetryService telemetry.TelemetryService
	clientId := config.GetClientId()
	telemetryEnabled := config.TelemetryEnabled()

	if telemetryEnabled {
		telemetryService = posthogservice.NewTelemetryService(posthogservice.PosthogServiceConfig{
			ApiKey:   internal.PosthogApiKey,
			Endpoint: internal.PosthogEndpoint,
		})
	}

	startTime := time.Now()

	cmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if telemetryService != nil {
			err := telemetryService.TrackCliEvent(telemetry.CliEventCmdStart, clientId, getCmdTelemetryData(cmd))
			if err != nil {
				log.Error(err)
			}
		}

		if output.FormatFlag == "" {
			return
		}
		originalStdout = os.Stdout
		os.Stdout = nil
	}

	cmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		endTime := time.Now()

		if telemetryService != nil {
			defer telemetryService.Close()
			execTime := endTime.Sub(startTime)
			props := getCmdTelemetryData(cmd)
			props["exec time (µs)"] = execTime.Microseconds()

			err := telemetryService.TrackCliEvent(telemetry.CliEventCmdEnd, clientId, props)
			if err != nil {
				log.Error(err)
			}
		}

		os.Stdout = originalStdout
		output.Print(output.Output, output.FormatFlag)
	}
}

func RunInitialScreenFlow(cmd *cobra.Command, args []string) {
	command, err := view.GetCommand()
	if err != nil {
		log.Fatal(err)
	}

	switch command {
	case "server":
		ServerCmd.Run(cmd, []string{})
	case "create":
		CreateCmd.Run(cmd, []string{})
	case "code":
		CodeCmd.Run(cmd, []string{})
	case "git-provider add":
		GitProviderAddCmd.Run(cmd, []string{})
	case "target set":
		TargetSetCmd.Run(cmd, []string{})
	case "docs":
		DocsCmd.Run(cmd, []string{})
	case "help":
		err := cmd.Help()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func getCmdTelemetryData(cmd *cobra.Command) map[string]interface{} {
	path := cmd.CommandPath()

	// Trim daytona from the path if a non-root command was invoked
	// This prevents a `daytona` pileup in the telemetry data
	if path != "daytona" {
		path = strings.TrimPrefix(path, "daytona ")
	}

	source := telemetry.CLI_SOURCE
	if internal.WorkspaceMode() {
		source = telemetry.CLI_PROJECT_SOURCE
	}

	calledAs := cmd.CalledAs()
	return map[string]interface{}{
		"command":   path,
		"called_as": calledAs,
		"source":    source,
	}
}
