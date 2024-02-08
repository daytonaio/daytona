// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"

	. "github.com/daytonaio/daytona/cmd/agent"
	. "github.com/daytonaio/daytona/cmd/ports"
	. "github.com/daytonaio/daytona/cmd/profile"
	. "github.com/daytonaio/daytona/cmd/workspace"
	"github.com/daytonaio/daytona/output"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "daytona",
	Short: "Daytona Agent",
	Long:  `Daytona Agent is a tool for managing development environments`,
}

var originalStdout *os.File

func Execute() {
	_, wsMode := os.LookupEnv("DAYTONA_WS_DIR")

	rootCmd.AddCommand(InfoCmd)
	rootCmd.AddCommand(StartCmd)
	rootCmd.AddCommand(StopCmd)
	rootCmd.AddCommand(exposePortCmd)
	rootCmd.AddCommand(PortsCmd)
	rootCmd.AddCommand(versionCmd)

	if !wsMode {
		rootCmd.AddCommand(CodeCmd)
		rootCmd.AddCommand(SshCmd)
		rootCmd.AddCommand(SshProxyCmd)
		rootCmd.AddCommand(CreateCmd)
		rootCmd.AddCommand(DeleteCmd)
		rootCmd.AddCommand(ListCmd)
		rootCmd.AddCommand(AgentCmd)
		rootCmd.AddCommand(ideCmd)
		rootCmd.AddCommand(ProfileCmd)
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
