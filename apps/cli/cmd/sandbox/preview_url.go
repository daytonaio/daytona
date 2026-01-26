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

var PreviewUrlCmd = &cobra.Command{
	Use:     "preview-url [SANDBOX_ID | SANDBOX_NAME]",
	Short:   "Get signed preview URL for a sandbox port",
	Args:    cobra.ExactArgs(1),
	Aliases: common.GetAliases("preview-url"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		sandboxIdOrName := args[0]

		if previewUrlPort == 0 {
			return fmt.Errorf("port flag is required")
		}

		req := apiClient.SandboxAPI.GetSignedPortPreviewUrl(ctx, sandboxIdOrName, previewUrlPort).
			ExpiresInSeconds(previewUrlExpires)

		previewUrl, res, err := req.Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		fmt.Println(previewUrl.Url)

		return nil
	},
}

var (
	previewUrlPort    int32
	previewUrlExpires int32
)

func init() {
	PreviewUrlCmd.Flags().Int32VarP(&previewUrlPort, "port", "p", 0, "Port number to get preview URL for (required)")
	PreviewUrlCmd.Flags().Int32Var(&previewUrlExpires, "expires", 3600, "URL expiration time in seconds")

	_ = PreviewUrlCmd.MarkFlagRequired("port")
}
