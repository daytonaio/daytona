// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/cmd/common"
	"github.com/daytonaio/daytona/cli/toolbox"
	apiclient_go "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/spf13/cobra"
)

var ExecCmd = &cobra.Command{
	Use:   "exec [SANDBOX_ID | SANDBOX_NAME] [flags] -- COMMAND [ARGS...]",
	Short: "Execute a command in a sandbox",
	Long:  "Execute a command in a running sandbox.\n\nFlags must be specified before -- which separates the sandbox identifier from the command to run.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		// cmd.ArgsLenAtDash() returns the number of args before "--", or -1 if "--" was not used.
		// Cobra consumes "--" itself, so we cannot search for it in args directly.
		dashIndex := cmd.ArgsLenAtDash()

		if dashIndex == -1 {
			return fmt.Errorf("use -- to separate the sandbox from the command: exec SANDBOX -- COMMAND [ARGS...]")
		}
		if dashIndex == 0 {
			return fmt.Errorf("sandbox ID or name is required: exec SANDBOX -- COMMAND [ARGS...]")
		}
		if dashIndex > 1 {
			return fmt.Errorf("unexpected arguments before --: only one sandbox ID or name is allowed")
		}

		sandboxIdOrName := args[0]
		commandArgs := args[dashIndex:]

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

		// If TTY mode is enabled, use interactive TTY execution
		if execTTY {
			return executeTTY(ctx, toolboxClient, sandbox, commandArgs)
		}

		// Otherwise use regular command execution
		return executeRegular(ctx, toolboxClient, sandbox, commandArgs)
	},
}

func executeRegular(ctx context.Context, toolboxClient *toolbox.Client, sandbox *apiclient_go.Sandbox, commandArgs []string) error {
	// Build command from args
	command := buildCommand(commandArgs)

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
}

func executeTTY(ctx context.Context, toolboxClient *toolbox.Client, sandbox *apiclient_go.Sandbox, commandArgs []string) error {
	sessionID := fmt.Sprintf("exec-%d", time.Now().UnixNano())

	executeRequest := toolbox.PTYCreateRequest{
		ID: sessionID,
	}

	if len(commandArgs) > 0 {
		cmd := commandArgs[0]
		executeRequest.Command = &cmd
		if len(commandArgs) > 1 {
			executeRequest.Args = commandArgs[1:]
		}
	}

	if execCwd != "" {
		executeRequest.Cwd = execCwd
	}
	if execTimeout > 0 {
		timeout := uint32(execTimeout)
		executeRequest.Timeout = &timeout
	}

	// Execute the command via PTY
	err := toolboxClient.ExecuteCommandTTY(ctx, sandbox, executeRequest)
	// If the remote process exited with a non-zero code, propagate it cleanly.
	// By the time we reach here, connectAndStreamTTY has already returned, so
	// defer term.Restore has run and the terminal is in its original state.
	var exitErr *toolbox.ExitCodeError
	if errors.As(err, &exitErr) {
		os.Exit(exitErr.Code)
	}
	return err
}

var (
	execCwd     string
	execTTY     bool
	execTimeout int
)

// quoteArg applies quoting rules so that parseCommand on the daemon side can
// re-split the command string back into the original argument slice.
func quoteArg(arg string) string {
	if strings.ContainsAny(arg, " \t'\"") {
		if !strings.Contains(arg, "'") {
			return "'" + arg + "'"
		}
		// Fall back to double-quote wrapping, escaping any internal double quotes.
		escaped := strings.ReplaceAll(arg, `"`, `\"`)
		return `"` + escaped + `"`
	}
	return arg
}

// buildCommand reconstructs the command string from args with proper quoting.
// For shell -c commands, only args[2] is treated as the script; any remaining
// args (which become $0, $1, … inside the script) are preserved as separate
// argv entries.
// For regular commands, arguments that contain whitespace or quotes are quoted
// so that parseCommand on the daemon side re-splits them correctly.
func buildCommand(args []string) string {
	if len(args) == 0 {
		return ""
	}

	// For shell -c commands, wrap only the script argument (args[2]) in single
	// quotes so that parseCommand on the daemon reconstructs the correct argv:
	//   [shell, "-c", script, arg0, arg1, ...]
	// Single quotes inside the script are escaped using the POSIX idiom: ' → '\''
	if len(args) >= 3 && args[1] == "-c" {
		cmdParts := make([]string, 0, len(args))
		cmdParts = append(cmdParts, args[0], args[1])

		// Script argument gets special single-quote wrapping.
		escapedScript := strings.ReplaceAll(args[2], "'", `'\''`)
		cmdParts = append(cmdParts, "'"+escapedScript+"'")

		// Additional arguments after the script are preserved as separate
		// argv entries, using the normal quoting rules.
		for _, arg := range args[3:] {
			cmdParts = append(cmdParts, quoteArg(arg))
		}

		return strings.Join(cmdParts, " ")
	}

	// For regular commands, quote arguments that contain whitespace or quote
	// characters so that parseCommand re-splits them correctly and does not
	// misinterpret unquoted quote characters as starting a quoted segment.
	quotedArgs := make([]string, len(args))
	for i, arg := range args {
		quotedArgs[i] = quoteArg(arg)
	}
	return strings.Join(quotedArgs, " ")
}

func init() {
	ExecCmd.Flags().StringVar(&execCwd, "cwd", "", "Working directory for command execution")
	ExecCmd.Flags().IntVar(&execTimeout, "timeout", 0, "Command timeout in seconds (0 for no timeout)")
	ExecCmd.Flags().BoolVar(&execTTY, "tty", false, "Enable TTY mode for interactive commands")
}
