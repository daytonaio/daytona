// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"slices"
	"strings"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	. "github.com/daytonaio/daytona/internal/util"
	. "github.com/daytonaio/daytona/pkg/cmd/apikey"
	. "github.com/daytonaio/daytona/pkg/cmd/autocomplete"
	. "github.com/daytonaio/daytona/pkg/cmd/build"
	cmd_common "github.com/daytonaio/daytona/pkg/cmd/common"
	. "github.com/daytonaio/daytona/pkg/cmd/env"
	. "github.com/daytonaio/daytona/pkg/cmd/gitprovider"
	. "github.com/daytonaio/daytona/pkg/cmd/ports"
	. "github.com/daytonaio/daytona/pkg/cmd/prebuild"
	. "github.com/daytonaio/daytona/pkg/cmd/profile"
	. "github.com/daytonaio/daytona/pkg/cmd/provider"
	. "github.com/daytonaio/daytona/pkg/cmd/runner"
	. "github.com/daytonaio/daytona/pkg/cmd/server"
	. "github.com/daytonaio/daytona/pkg/cmd/target"
	. "github.com/daytonaio/daytona/pkg/cmd/targetconfig"
	. "github.com/daytonaio/daytona/pkg/cmd/telemetry"
	. "github.com/daytonaio/daytona/pkg/cmd/workspace"
	. "github.com/daytonaio/daytona/pkg/cmd/workspace/create"
	. "github.com/daytonaio/daytona/pkg/cmd/workspacetemplate"
	"github.com/daytonaio/daytona/pkg/common"
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
	rootCmd.AddGroup(&cobra.Group{ID: TARGET_GROUP, Title: "Targets & Workspaces"})
	rootCmd.AddGroup(&cobra.Group{ID: SERVER_GROUP, Title: "Server"})
	rootCmd.AddGroup(&cobra.Group{ID: PROFILE_GROUP, Title: "Profile"})
	rootCmd.AddGroup(&cobra.Group{ID: RUNNER_GROUP, Title: "Runner"})

	rootCmd.AddCommand(CodeCmd)
	rootCmd.AddCommand(SshCmd)
	rootCmd.AddCommand(SshProxyCmd)
	rootCmd.AddCommand(CreateCmd)
	rootCmd.AddCommand(DeleteCmd)
	rootCmd.AddCommand(WorkspaceTemplateCmd)
	rootCmd.AddCommand(ServeCmd)
	rootCmd.AddCommand(DaemonServeCmd)
	rootCmd.AddCommand(ServerCmd)
	rootCmd.AddCommand(ApiKeyCmd)
	rootCmd.AddCommand(ProviderCmd)
	rootCmd.AddCommand(TargetConfigCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(ideCmd)
	rootCmd.AddCommand(RunnerCmd)
	rootCmd.AddCommand(ProfileCmd)
	rootCmd.AddCommand(ProfileUseCmd)
	rootCmd.AddCommand(whoamiCmd)
	rootCmd.AddCommand(purgeCmd)
	rootCmd.AddCommand(GitProviderCmd)
	rootCmd.AddCommand(StartCmd)
	rootCmd.AddCommand(StopCmd)
	rootCmd.AddCommand(RestartCmd)
	rootCmd.AddCommand(LogsCmd)
	rootCmd.AddCommand(InfoCmd)
	rootCmd.AddCommand(PrebuildCmd)
	rootCmd.AddCommand(BuildCmd)
	rootCmd.AddCommand(PortForwardCmd)
	rootCmd.AddCommand(EnvCmd)
	rootCmd.AddCommand(TelemetryCmd)
	rootCmd.AddCommand(TargetCmd)

	SetupRootCommand(rootCmd)

	startTime := time.Now()
	clientId := config.GetClientId()
	telemetryEnabled := config.TelemetryEnabled()

	cmd, flags, isCompletion, err := PreRun(rootCmd, os.Args[1:], telemetryEnabled, clientId, startTime)
	if err != nil {
		fmt.Printf("Error: %v\n\n", err)
		return cmd.Help()
	}

	err = rootCmd.Execute()

	endTime := time.Now()

	if !isCompletion {
		PostRun(cmd, err, clientId, startTime, endTime, flags)
	}

	return err
}

func validateCommands(rootCmd *cobra.Command, args []string) (cmd *cobra.Command, flags []string, isCompletion bool, err error) {
	completionCommands := []string{"__complete", "__completeNoDesc", "__completeNoDescCmd", "__completeCmd"}
	if len(args) > 0 && slices.Contains(completionCommands, args[0]) {
		return rootCmd, flags, true, nil
	}

	rootCmd.InitDefaultHelpCmd()
	currentCmd := rootCmd

	// Filter flags from args
	sanitizedArgs := []string{}
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "-") {
			flags = append(flags, args[i])
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
			}
			continue
		}
		sanitizedArgs = append(sanitizedArgs, args[i])
	}

	for len(sanitizedArgs) > 0 {
		subCmd, subArgs, err := currentCmd.Find(sanitizedArgs)
		if err != nil {
			return currentCmd, flags, false, err
		}

		if subCmd == currentCmd {
			break
		}

		currentCmd = subCmd
		sanitizedArgs = subArgs
	}

	return currentCmd, flags, false, currentCmd.ValidateArgs(sanitizedArgs)
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
	case "target-config set":
		return TargetConfigAddCmd.RunE(cmd, []string{})
	case "docs":
		return DocsCmd.RunE(cmd, []string{})
	case "help":
		return cmd.Help()
	}

	return nil
}

func PreRun(rootCmd *cobra.Command, args []string, telemetryEnabled bool, clientId string, startTime time.Time) (*cobra.Command, []string, bool, error) {
	cmd, flags, isCompletion, err := validateCommands(rootCmd, os.Args[1:])
	if err != nil && !isCompletion {
		if !shouldIgnoreCommand(cmd.CommandPath()) {
			event := telemetry.NewCliEvent(telemetry.CliEventCommandInvalid, cmd, flags, err, nil)
			err := cmd_common.TrackTelemetryEvent(event, clientId)
			if err != nil {
				log.Trace(err)
			}
			err = cmd_common.CloseTelemetryService()
			if err != nil {
				log.Trace(err)
			}
		}

		return cmd, flags, isCompletion, err
	}

	if !shouldIgnoreCommand(cmd.CommandPath()) && !isCompletion {
		event := telemetry.NewCliEvent(telemetry.CliEventCommandStarted, cmd, flags, nil, nil)
		err := cmd_common.TrackTelemetryEvent(event, clientId)
		if err != nil {
			log.Trace(err)
		}

		go func() {
			interruptChannel := make(chan os.Signal, 1)
			signal.Notify(interruptChannel, os.Interrupt)

			for range interruptChannel {
				endTime := time.Now()
				execTime := endTime.Sub(startTime)
				extras := map[string]interface{}{"exec_time_µs": execTime.Microseconds()}
				event := telemetry.NewCliEvent(telemetry.CliEventCommandInterrupted, cmd, flags, nil, extras)
				err := cmd_common.TrackTelemetryEvent(event, clientId)
				if err != nil {
					log.Trace(err)
				}

				err = cmd_common.CloseTelemetryService()
				if err != nil {
					log.Trace(err)
				}
				os.Exit(0)
			}
		}()
	}

	return cmd, flags, isCompletion, nil
}

func PostRun(cmd *cobra.Command, cmdErr error, clientId string, startTime time.Time, endTime time.Time, flags []string) {
	if !shouldIgnoreCommand(cmd.CommandPath()) {
		execTime := endTime.Sub(startTime)
		extras := map[string]interface{}{"exec_time_µs": execTime.Microseconds()}
		eventName := telemetry.CliEventCommandCompleted
		if cmdErr != nil {
			eventName = telemetry.CliEventCommandFailed
		}
		event := telemetry.NewCliEvent(eventName, cmd, flags, cmdErr, extras)

		err := cmd_common.TrackTelemetryEvent(event, clientId)
		if err != nil {
			log.Trace(err)
		}

		err = cmd_common.CloseTelemetryService()
		if err != nil {
			log.Trace(err)
		}
	}
}

func shouldIgnoreCommand(commandPath string) bool {
	ignoredPaths := []string{"daemon-serve", "ssh-proxy"}

	for _, ignoredPath := range ignoredPaths {
		if strings.HasSuffix(commandPath, ignoredPath) {
			return true
		}
	}

	return false
}
