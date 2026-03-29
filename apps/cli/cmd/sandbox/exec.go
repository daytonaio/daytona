// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/cmd/common"
	"github.com/daytonaio/daytona/cli/toolbox"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var ExecCmd = &cobra.Command{
	Use:   "exec [SANDBOX_ID | SANDBOX_NAME] -- [COMMAND] [ARGS...]",
	Short: "Execute a command in a sandbox",
	Long: `Execute a command in a running sandbox

Examples:
  # Execute a simple command
  daytona sandbox exec my-sandbox -- ls -la

  # Execute an interactive command with TTY
  daytona sandbox exec my-sandbox -t -- bash

  # Execute with custom working directory
  daytona sandbox exec my-sandbox --cwd /app -- npm start

  # Execute with timeout (non-TTY mode only)
  daytona sandbox exec my-sandbox --timeout 30 -- long-running-command`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		sandboxIdOrName := args[0]

		// Find the command args after "--"
		commandArgs := args[1:]
		if len(commandArgs) == 0 {
			return fmt.Errorf("no command specified")
		}

		// First, get the sandbox to get its ID and region (in case name was provided)
		sandbox, res, err := apiClient.SandboxAPI.GetSandbox(ctx, sandboxIdOrName).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		if err := common.RequireStartedState(sandbox); err != nil {
			return err
		}

		toolboxClient := toolbox.NewClient(apiClient)

		command := strings.Join(commandArgs, " ")

		// Use TTY mode if requested
		if execTTY {
			// Validate TTY mode requirements
			if !term.IsTerminal(int(os.Stdin.Fd())) {
				return fmt.Errorf("--tty flag requires an interactive terminal (stdin is not a terminal)")
			}

			// TTY mode is not compatible with timeout
			if execTimeout > 0 {
				return fmt.Errorf("--timeout and --tty flags cannot be used together (TTY sessions are inherently interactive)")
			}

			// Get terminal size for TTY mode
			cols, rows := 80, 24 // Default values
			if c, r, err := term.GetSize(int(os.Stdin.Fd())); err == nil {
				cols, rows = c, r
			}

			ttyRequest := toolbox.TTYExecuteRequest{
				Command: command,
				Cols:    cols,
				Rows:    rows,
			}
			if execCwd != "" {
				ttyRequest.Cwd = &execCwd
			}

			// Execute with TTY (WebSocket) - this handles the entire session
			return toolboxClient.ExecuteCommandWithTTY(ctx, sandbox, ttyRequest)
		}

		// Regular execution mode (non-TTY)
		executeRequest := toolbox.ExecuteRequest{
			Command: command,
		}
		if execCwd != "" {
			executeRequest.Cwd = &execCwd
		}
		if execTimeout > 0 {
			timeout := float32(execTimeout)
			executeRequest.Timeout = &timeout
		}

		// Execute the command via toolbox
		response, err := toolboxClient.ExecuteCommand(ctx, sandbox, executeRequest)
		if err != nil {
			return err
		}

		// Print the output (stdout + stderr combined)
		if response.Result != "" {
			fmt.Print(response.Result)
		}

		// Exit with the command's exit code
		exitCode := int(response.ExitCode)
		if exitCode != 0 {
			if response.Result == "" {
				fmt.Fprintf(os.Stderr, "Command failed with exit code %d\n", exitCode)
			}
			os.Exit(exitCode)
		}

		return nil
	},
}

var (
	execCwd     string
	execTimeout int
	execTTY     bool
)

func init() {
	ExecCmd.Flags().StringVar(&execCwd, "cwd", "", "Working directory for command execution")
	ExecCmd.Flags().IntVar(&execTimeout, "timeout", 0, "Command timeout in seconds (0 for no timeout)")
	ExecCmd.Flags().BoolVarP(&execTTY, "tty", "t", false, "Allocate a pseudo-TTY for interactive command execution")
}
