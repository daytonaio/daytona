// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/cli/cmd/common"
	"github.com/daytonaio/daytona/cli/config"
	"github.com/spf13/cobra"
)

var LogsCmd = &cobra.Command{
	Use:     "logs [SANDBOX_ID] | [SANDBOX_NAME]",
	Short:   "Get sandbox logs",
	Args:    cobra.ExactArgs(1),
	Aliases: common.GetAliases("logs"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		sandboxIdOrNameArg := args[0]
		showTimestamps, _ := cmd.Flags().GetBool("timestamps")

		// Get config to access server URL and auth
		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			return err
		}

		// Build URL with timestamps parameter
		url := fmt.Sprintf("%s/sandbox/%s/logs", activeProfile.Api.Url, sandboxIdOrNameArg)
		if showTimestamps {
			url += "?timestamps=true"
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %v", err)
		}

		// Add authorization header
		if activeProfile.Api.Key != nil {
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", *activeProfile.Api.Key))
		} else if activeProfile.Api.Token != nil {
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", activeProfile.Api.Token.AccessToken))
			if activeProfile.ActiveOrganizationId != nil {
				req.Header.Add("X-Daytona-Organization-ID", *activeProfile.ActiveOrganizationId)
			}
		}

		req.Header.Add("Accept", "text/plain")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to connect to server: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("server returned a non-OK status while retrieving logs: %d", resp.StatusCode)
		}

		// Read and print the response body
		buf := make([]byte, 4096)
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				fmt.Print(string(buf[:n]))
			}
			if err != nil {
				break
			}
		}

		return nil
	},
}

func init() {
	LogsCmd.Flags().Bool("timestamps", false, "Show timestamps in logs (default: false)")
	SandboxCmd.AddCommand(LogsCmd)
}
