// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"

	. "github.com/daytonaio/daytona/pkg/cmd/agent"
	. "github.com/daytonaio/daytona/pkg/cmd/gitprovider"
	"github.com/daytonaio/daytona/pkg/cmd/output"
	. "github.com/daytonaio/daytona/pkg/cmd/ports"
	. "github.com/daytonaio/daytona/pkg/cmd/profile"
	. "github.com/daytonaio/daytona/pkg/cmd/provider"
	. "github.com/daytonaio/daytona/pkg/cmd/server"
	. "github.com/daytonaio/daytona/pkg/cmd/workspace"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "daytona",
	Short: "Daytona Server",
	Long:  `Daytona Server is a tool for managing development environments`,
}

var originalStdout *os.File

func Execute() {
	_, wsMode := os.LookupEnv("DAYTONA_WS_DIR")

	rootCmd.AddCommand(InfoCmd)
	rootCmd.AddCommand(StartCmd)
	rootCmd.AddCommand(StopCmd)
	rootCmd.AddCommand(PortsCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(ListCmd)
	rootCmd.AddCommand(GitProviderCmd)

	if wsMode {
		rootCmd.AddCommand(gitCredCmd)
		rootCmd.AddCommand(AgentCmd)
	} else {
		rootCmd.AddCommand(CodeCmd)
		rootCmd.AddCommand(SshCmd)
		rootCmd.AddCommand(SshProxyCmd)
		rootCmd.AddCommand(CreateCmd)
		rootCmd.AddCommand(DeleteCmd)
		rootCmd.AddCommand(ServerCmd)
		rootCmd.AddCommand(ideCmd)
		rootCmd.AddCommand(ProfileCmd)
		rootCmd.AddCommand(ProviderCmd)
	}

	rootCmd.PersistentFlags().BoolP("help", "", false, "help for daytona")
	rootCmd.PersistentFlags().StringVarP(&output.FormatFlag, "output", "o", output.FormatFlag, `Output format. Must be one of (yaml, json)`)

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if output.FormatFlag == "" {
			return
		}
		originalStdout = os.Stdout
		os.Stdout = nil
	}

	rootCmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		os.Stdout = originalStdout
		output.Print(output.Output, output.FormatFlag)
	}

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
