// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/daytonaio/daytona/cli/cmd"
	"github.com/daytonaio/daytona/cli/cmd/auth"
	"github.com/daytonaio/daytona/cli/cmd/common"
	"github.com/daytonaio/daytona/cli/cmd/mcp"
	"github.com/daytonaio/daytona/cli/cmd/organization"
	"github.com/daytonaio/daytona/cli/cmd/sandbox"
	"github.com/daytonaio/daytona/cli/cmd/snapshot"
	"github.com/daytonaio/daytona/cli/cmd/volume"
	"github.com/daytonaio/daytona/cli/internal"
	"github.com/daytonaio/daytona/cli/internal/clierr"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "daytona",
	Short: "Daytona CLI",
	Long: `Command line interface for Daytona Sandboxes

Exit codes: 0 success; 1 runtime failure; 2 invalid flags or arguments (where validated); 124 wait timeout. 'daytona exec' exits with the remote command's exit code; 255 indicates a CLI-side failure.`,
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

	// Add sandbox subcommands as top-level shortcuts
	rootCmd.AddCommand(createSandboxShortcut(sandbox.CreateCmd))
	rootCmd.AddCommand(createSandboxShortcut(sandbox.DeleteCmd))
	rootCmd.AddCommand(createSandboxShortcut(sandbox.InfoCmd))
	rootCmd.AddCommand(createSandboxShortcut(sandbox.ListCmd))
	rootCmd.AddCommand(createSandboxShortcut(sandbox.StartCmd))
	rootCmd.AddCommand(createSandboxShortcut(sandbox.StopCmd))
	rootCmd.AddCommand(createSandboxShortcut(sandbox.ArchiveCmd))
	rootCmd.AddCommand(createSandboxShortcut(sandbox.SSHCmd))
	rootCmd.AddCommand(createSandboxShortcut(sandbox.ExecCmd))
	rootCmd.AddCommand(createSandboxShortcut(sandbox.PreviewUrlCmd))

	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.PersistentFlags().BoolP("help", "", false, "help for daytona")
	rootCmd.PersistentFlags().BoolVar(&internal.NoInput, "no-input", false, "Never prompt for input; fail instead when input would be required")
	rootCmd.Flags().BoolP("version", "v", false, "Display the version of Daytona")

	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		return clierr.New(clierr.CategoryUsage, err.Error())
	})

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

// createSandboxShortcut creates a top-level shortcut for a sandbox subcommand
func createSandboxShortcut(original *cobra.Command) *cobra.Command {
	shortcut := &cobra.Command{
		Use:     original.Use,
		Short:   original.Short,
		Long:    original.Long,
		Example: original.Example,
		Args:    original.Args,
		Aliases: original.Aliases,
		GroupID: internal.SANDBOX_GROUP,
		PreRunE: original.PreRunE,
		RunE:    original.RunE,
	}
	shortcut.Flags().AddFlagSet(original.Flags())
	return shortcut
}

func main() {
	_ = godotenv.Load()

	err := rootCmd.Execute()
	if err != nil {
		reportExecuteError(err)
		os.Exit(clierr.ExitCode(err))
	}
}

// reportExecuteError prints the final error: a single-line JSON object on
// stderr when a structured output format was requested, a human-readable log
// line otherwise.
func reportExecuteError(err error) {
	if common.FormatFlag == "" {
		log.Error(err)
		return
	}

	payload := struct {
		Error string `json:"error"`
		Code  string `json:"code"`
		Hint  string `json:"hint,omitempty"`
	}{Error: err.Error(), Code: "error"}

	var cliErr *clierr.Error
	if errors.As(err, &cliErr) {
		payload.Error = cliErr.Message
		payload.Code = string(cliErr.Category)
		payload.Hint = cliErr.Hint
	}

	data, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		log.Error(err)
		return
	}
	fmt.Fprintln(os.Stderr, string(data))
}
