// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	. "dagent/cmd/agent"
	. "dagent/cmd/ports"
	. "dagent/cmd/profile"
	. "dagent/cmd/workspace"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "daytona",
	Short: "Daytona Agent",
	Long:  `Daytona Agent is a tool for managing development environments`,
}

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

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
