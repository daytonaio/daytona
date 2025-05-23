// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/daytonaio/daytona/cli/cmd"
	"github.com/daytonaio/daytona/cli/cmd/auth"
	"github.com/daytonaio/daytona/cli/cmd/mcp"
	"github.com/daytonaio/daytona/cli/cmd/organization"
	"github.com/daytonaio/daytona/cli/cmd/sandbox"
	"github.com/daytonaio/daytona/cli/cmd/snapshot"
	"github.com/daytonaio/daytona/cli/cmd/volume"
	"github.com/daytonaio/daytona/cli/internal"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:               "daytona",
	Short:             "Daytona CLI",
	Long:              "Command line interface for Daytona Sandboxes",
	DisableAutoGenTag: true,
	SilenceUsage:      true,
	SilenceErrors:     true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCmd.AddGroup(&cobra.Group{ID: internal.USER_GROUP, Title: "User"})
	rootCmd.AddGroup(&cobra.Group{ID: internal.SANDBOX_GROUP, Title: "Sandbox"})

	rootCmd.AddCommand(auth.LoginCmd)
	rootCmd.AddCommand(auth.LogoutCmd)
	rootCmd.AddCommand(sandbox.SandboxCmd)
	rootCmd.AddCommand(snapshot.SnapshotsCmd)
	rootCmd.AddCommand(volume.VolumeCmd)
	rootCmd.AddCommand(organization.OrganizationCmd)
	rootCmd.AddCommand(mcp.MCPCmd)
	rootCmd.AddCommand(cmd.DocsCmd)
	rootCmd.AddCommand(cmd.AutoCompleteCmd)
	rootCmd.AddCommand(cmd.GenerateDocsCmd)
	rootCmd.AddCommand(cmd.VersionCmd)

	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.PersistentFlags().BoolP("help", "", false, "help for daytona")
	rootCmd.Flags().BoolP("version", "v", false, "Display the version of Daytona")

	rootCmd.PreRun = func(command *cobra.Command, args []string) {
		versionFlag, _ := command.Flags().GetBool("version")
		if versionFlag {
			err := cmd.VersionCmd.RunE(command, []string{})
			if err != nil {
				log.Fatal(err)
			}
			os.Exit(0)
		}
	}
}

func main() {
	_ = godotenv.Load()

	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
