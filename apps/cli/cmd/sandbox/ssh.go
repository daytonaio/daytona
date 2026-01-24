// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/cmd/common"
	"github.com/spf13/cobra"
)

var SSHCmd = &cobra.Command{
	Use:   "ssh [SANDBOX_ID] | [SANDBOX_NAME]",
	Short: "SSH into a sandbox",
	Long:  "Establish an SSH connection to a running sandbox",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		sandboxIdOrName := args[0]

		// Get sandbox to check state
		sandbox, res, err := apiClient.SandboxAPI.GetSandbox(ctx, sandboxIdOrName).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		if err := common.RequireStartedState(sandbox); err != nil {
			return err
		}

		// Create SSH access token
		sshAccessRequest := apiClient.SandboxAPI.CreateSshAccess(ctx, sandbox.Id)
		if sshExpiresInMinutes > 0 {
			sshAccessRequest = sshAccessRequest.ExpiresInMinutes(float32(sshExpiresInMinutes))
		}

		sshAccess, res, err := sshAccessRequest.Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		// Parse the SSH command from the response
		sshArgs, err := common.ParseSSHCommand(sshAccess.SshCommand)
		if err != nil {
			return fmt.Errorf("failed to parse SSH command: %w", err)
		}

		// Execute SSH
		return common.ExecuteSSH(sshArgs)
	},
}

var sshExpiresInMinutes int

func init() {
	SSHCmd.Flags().IntVar(&sshExpiresInMinutes, "expires", 1440, "SSH access token expiration time in minutes (defaults to 24 hours)")
}
