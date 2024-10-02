// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal"
	. "github.com/daytonaio/daytona/internal/util"
	. "github.com/daytonaio/daytona/pkg/cmd/apikey"
	. "github.com/daytonaio/daytona/pkg/cmd/build"
	. "github.com/daytonaio/daytona/pkg/cmd/containerregistry"
	. "github.com/daytonaio/daytona/pkg/cmd/gitprovider"
	. "github.com/daytonaio/daytona/pkg/cmd/ports"
	. "github.com/daytonaio/daytona/pkg/cmd/prebuild"
	. "github.com/daytonaio/daytona/pkg/cmd/profile"
	. "github.com/daytonaio/daytona/pkg/cmd/profiledata/env"
	. "github.com/daytonaio/daytona/pkg/cmd/projectconfig"
	. "github.com/daytonaio/daytona/pkg/cmd/provider"
	. "github.com/daytonaio/daytona/pkg/cmd/server"
	. "github.com/daytonaio/daytona/pkg/cmd/target"
	. "github.com/daytonaio/daytona/pkg/cmd/telemetry"
	. "github.com/daytonaio/daytona/pkg/cmd/workspace"
	"github.com/daytonaio/daytona/pkg/common"
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
	SilenceUsage:      true,
	SilenceErrors:     true,
	DisableAutoGenTag: true,
	RunE:              RunInitialScreenFlow,
}

func Execute() error {
	rootCmd.AddGroup(&cobra.Group{ID: WORKSPACE_GROUP, Title: "Workspaces & Projects"})
	rootCmd.AddGroup(&cobra.Group{ID: SERVER_GROUP, Title: "Server"})
	rootCmd.AddGroup(&cobra.Group{ID: PROFILE_GROUP, Title: "Profile"})

	rootCmd.AddCommand(CodeCmd)
	rootCmd.AddCommand(SshCmd)
	rootCmd.AddCommand(SshProxyCmd)
	rootCmd.AddCommand(CreateCmd)
	rootCmd.AddCommand(DeleteCmd)
	rootCmd.AddCommand(ProjectConfigCmd)
	rootCmd.AddCommand(ServeCmd)
	rootCmd.AddCommand(ServerCmd)
	rootCmd.AddCommand(ApiKeyCmd)
	rootCmd.AddCommand(ContainerRegistryCmd)
	rootCmd.AddCommand(ProviderCmd)
	rootCmd.AddCommand(TargetCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(ideCmd)
	rootCmd.AddCommand(ProfileCmd)
	rootCmd.AddCommand(ProfileUseCmd)
	rootCmd.AddCommand(whoamiCmd)
	rootCmd.AddCommand(purgeCmd)
	rootCmd.AddCommand(GitProviderCmd)
	rootCmd.AddCommand(StartCmd)
	rootCmd.AddCommand(StopCmd)
	rootCmd.AddCommand(RestartCmd)
	rootCmd.AddCommand(InfoCmd)
	rootCmd.AddCommand(PrebuildCmd)
	rootCmd.AddCommand(BuildCmd)
	rootCmd.AddCommand(PortForwardCmd)
	rootCmd.AddCommand(EnvCmd)
	rootCmd.AddCommand(TelemetryCmd)

	SetupRootCommand(rootCmd)

	startTime := time.Now()
	clientId := config.GetClientId()
	telemetryEnabled := config.TelemetryEnabled()

	telemetryService, cmd, flags, err := PreRun(rootCmd, os.Args[1:], telemetryEnabled, clientId, startTime)
	if err != nil {
		fmt.Printf("Error: %v\n\n", err)
		return cmd.Help()
	}

	err = rootCmd.Execute()

	endTime := time.Now()

	PostRun(cmd, err, telemetryService, clientId, startTime, endTime, flags)

	return err
}

func validateCommands(rootCmd *cobra.Command, args []string) (cmd *cobra.Command, flags []string, err error) {
	rootCmd.InitDefaultHelpCmd()
	currentCmd := rootCmd

	// Filter flags from args
	sanitzedArgs := []string{}
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "-") {
			flags = append(flags, args[i])
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
			}
			continue
		}
		sanitzedArgs = append(sanitzedArgs, args[i])
	}

	for len(sanitzedArgs) > 0 {
		subCmd, subArgs, err := currentCmd.Find(sanitzedArgs)
		if err != nil {
			return currentCmd, flags, err
		}

		if subCmd == currentCmd {
			break
		}

		currentCmd = subCmd
		sanitzedArgs = subArgs
	}

	if len(sanitzedArgs) > 0 {
		if err := currentCmd.ValidateArgs(sanitzedArgs); err != nil {
			return currentCmd, flags, err
		}
	}

	return currentCmd, flags, nil
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
	cmd.Flags().BoolP("version", "v", false, "Display the version of Daytona")

	cmd.PreRun = func(cmd *cobra.Command, args []string) {
		versionFlag, _ := cmd.Flags().GetBool("version")
		if versionFlag {
			err := versionCmd.RunE(cmd, []string{})
			if err != nil {
				log.Fatal(err)
			}
			os.Exit(0)
		}
	}
}

func RunInitialScreenFlow(cmd *cobra.Command, args []string) error {
	command, err := view.GetCommand()
	if err != nil {
		if common.IsCtrlCAbort(err) {
			return nil
		} else {
			return err
		}
	}

	switch command {
	case "server":
		return ServerCmd.RunE(cmd, []string{})
	case "create":
		return CreateCmd.RunE(cmd, []string{})
	case "code":
		return CodeCmd.RunE(cmd, []string{})
	case "git-provider add":
		return GitProviderAddCmd.RunE(cmd, []string{})
	case "target set":
		return TargetSetCmd.RunE(cmd, []string{})
	case "docs":
		return DocsCmd.RunE(cmd, []string{})
	case "help":
		return cmd.Help()
	}

	return nil
}

func GetCmdTelemetryData(cmd *cobra.Command, flags []string) map[string]interface{} {
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

	data := telemetry.AdditionalData
	data["command"] = path
	data["called_as"] = calledAs
	data["source"] = source
	data["flags"] = flags

	return data
}

func PreRun(rootCmd *cobra.Command, args []string, telemetryEnabled bool, clientId string, startTime time.Time) (telemetry.TelemetryService, *cobra.Command, []string, error) {
	var telemetryService telemetry.TelemetryService

	if telemetryEnabled {
		telemetryService = posthogservice.NewTelemetryService(posthogservice.PosthogServiceConfig{
			ApiKey:   internal.PosthogApiKey,
			Endpoint: internal.PosthogEndpoint,
			Version:  internal.Version,
		})
	}

	cmd, flags, err := validateCommands(rootCmd, os.Args[1:])
	if err != nil {
		if telemetryEnabled {
			props := GetCmdTelemetryData(cmd, flags)
			props["command"] = os.Args[1]
			props["called_as"] = os.Args[1]
			err := telemetryService.TrackCliEvent(telemetry.CliEventInvalidCmd, clientId, props)
			if err != nil {
				log.Error(err)
			}
			telemetryService.Close()
		}

		return telemetryService, cmd, flags, err
	}

	if telemetryEnabled {
		err := telemetryService.TrackCliEvent(telemetry.CliEventCmdStart, clientId, GetCmdTelemetryData(cmd, flags))
		if err != nil {
			log.Error(err)
		}

		go func() {
			interruptChannel := make(chan os.Signal, 1)
			signal.Notify(interruptChannel, os.Interrupt)

			for range interruptChannel {
				endTime := time.Now()
				execTime := endTime.Sub(startTime)
				props := GetCmdTelemetryData(cmd, flags)
				props["exec time (µs)"] = execTime.Microseconds()
				props["error"] = "interrupted"

				err := telemetryService.TrackCliEvent(telemetry.CliEventCmdEnd, clientId, props)
				if err != nil {
					log.Error(err)
				}
				telemetryService.Close()
				os.Exit(0)
			}
		}()
	}

	return telemetryService, cmd, flags, nil
}

func PostRun(cmd *cobra.Command, cmdErr error, telemetryService telemetry.TelemetryService, clientId string, startTime time.Time, endTime time.Time, flags []string) {
	if telemetryService != nil {
		execTime := endTime.Sub(startTime)
		props := GetCmdTelemetryData(cmd, flags)
		props["exec time (µs)"] = execTime.Microseconds()
		if cmdErr != nil {
			props["error"] = cmdErr.Error()
		}

		err := telemetryService.TrackCliEvent(telemetry.CliEventCmdEnd, clientId, props)
		if err != nil {
			log.Error(err)
		}
		telemetryService.Close()
	}
}
