// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/cmd/common"
	"github.com/daytonaio/daytona/cli/internal/clierr"
	"github.com/daytonaio/daytona/cli/toolbox"
	"github.com/spf13/cobra"
)

// execResult is the structured output of `daytona exec` in --format mode.
type execResult struct {
	Result   string `json:"result" yaml:"result"`
	ExitCode int    `json:"exitCode" yaml:"exitCode"`
}

var ExecCmd = &cobra.Command{
	Use:   "exec [SANDBOX_ID | SANDBOX_NAME] -- [COMMAND] [ARGS...]",
	Short: "Execute a command in a sandbox",
	Long: `Execute a command in a running sandbox.

Exits with the remote command's exit code; exit code 255 indicates a CLI-side failure (for example the sandbox was not found or is not running).`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient.GetApiClient(nil, nil)
		if err != nil {
			return execFailure(err)
		}

		sandboxIdOrName := args[0]

		// The command and its arguments follow the sandbox reference (after "--").
		commandArgs := args[1:]

		// First, get the sandbox to get its ID and region (in case name was provided)
		sandbox, res, err := apiClient.SandboxAPI.GetSandbox(ctx, sandboxIdOrName).Execute()
		if err != nil {
			return execFailure(apiclient.HandleErrorResponse(res, err))
		}

		if err := common.RequireStartedState(sandbox); err != nil {
			return execFailure(err)
		}

		toolboxClient := toolbox.NewClient(apiClient)

		command := common.ShellJoinArgs(commandArgs)

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
			return execFailure(err)
		}

		exitCode := int(response.ExitCode)

		if common.FormatFlag != "" {
			common.NewFormatter(execResult{
				Result:   response.Result,
				ExitCode: exitCode,
			}).Print()
			if exitCode != 0 {
				os.Exit(exitCode)
			}
			return nil
		}

		// Print the output (stdout + stderr combined)
		if response.Result != "" {
			fmt.Print(response.Result)
		}

		// Exit with the command's exit code
		if exitCode != 0 {
			if response.Result == "" {
				fmt.Fprintf(os.Stderr, "Command failed with exit code %d\n", exitCode)
			}
			os.Exit(exitCode)
		}

		return nil
	},
}

// execFailure marks a CLI-side failure (as opposed to a non-zero exit of the
// remote command) so `daytona exec` exits with 255 per the documented
// exit-code contract. A *clierr.Error keeps its category and hint; other
// errors are wrapped as server errors.
func execFailure(err error) error {
	var cliErr *clierr.Error
	if errors.As(err, &cliErr) {
		return cliErr.WithCode(255)
	}
	return clierr.New(clierr.CategoryServer, err.Error()).WithCode(255)
}

var (
	execCwd     string
	execTimeout int
)

func init() {
	ExecCmd.Flags().StringVar(&execCwd, "cwd", "", "Working directory for command execution")
	ExecCmd.Flags().IntVar(&execTimeout, "timeout", 0, "Command timeout in seconds (0 for no timeout)")
	common.RegisterFormatFlag(ExecCmd)
}
