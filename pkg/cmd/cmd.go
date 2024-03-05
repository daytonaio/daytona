// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"

	"github.com/daytonaio/daytona/internal/util"
	. "github.com/daytonaio/daytona/pkg/cmd/agent"
	. "github.com/daytonaio/daytona/pkg/cmd/gitprovider"
	"github.com/daytonaio/daytona/pkg/cmd/output"
	. "github.com/daytonaio/daytona/pkg/cmd/ports"
	. "github.com/daytonaio/daytona/pkg/cmd/profile"
	. "github.com/daytonaio/daytona/pkg/cmd/provider"
	. "github.com/daytonaio/daytona/pkg/cmd/server"
	. "github.com/daytonaio/daytona/pkg/cmd/target"
	. "github.com/daytonaio/daytona/pkg/cmd/workspace"
	view_util "github.com/daytonaio/daytona/pkg/views/util"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "daytona",
	Short: "Daytona is a Dev Environment Manager",
	Long:  "Daytona is a Dev Environment Manager",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(view_util.GetLongDescription())
		cmd.Help()
	},
}

var originalStdout *os.File

func Execute() {
	rootCmd.AddCommand(InfoCmd)
	rootCmd.AddCommand(StartCmd)
	rootCmd.AddCommand(StopCmd)
	rootCmd.AddCommand(PortForwardCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(ListCmd)
	rootCmd.AddCommand(GitProviderCmd)
	rootCmd.AddCommand(TargetCmd)

	if util.WorkspaceMode() {
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
		rootCmd.AddCommand(ProfileUseCmd)
		rootCmd.AddCommand(whoamiCmd)
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
