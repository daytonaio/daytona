// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"fmt"

	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/cmd/common"
	"github.com/daytonaio/daytona/cli/internal/clierr"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/spf13/cobra"
)

// previewUrlOutput is the structured-output shape of the preview-url command.
type previewUrlOutput struct {
	Url              string `json:"url" yaml:"url"`
	Port             int32  `json:"port" yaml:"port"`
	ExpiresInSeconds int32  `json:"expiresInSeconds" yaml:"expiresInSeconds"`
	Sandbox          string `json:"sandbox" yaml:"sandbox"`
}

// newPreviewUrlOutput assembles the structured output for a signed preview
// URL; sandbox echoes the user-supplied sandbox ID or name argument.
func newPreviewUrlOutput(previewUrl *apiclient.SignedPortPreviewUrl, expiresInSeconds int32, sandbox string) previewUrlOutput {
	return previewUrlOutput{
		Url:              previewUrl.Url,
		Port:             previewUrl.Port,
		ExpiresInSeconds: expiresInSeconds,
		Sandbox:          sandbox,
	}
}

var PreviewUrlCmd = &cobra.Command{
	Use:   "preview-url [SANDBOX_ID | SANDBOX_NAME]",
	Short: "Get signed preview URL for a sandbox port",
	Example: `  daytona preview-url my-sandbox --port 3000
  daytona preview-url my-sandbox --port 3000 --expires 7200
  daytona preview-url my-sandbox --port 3000 --format json`,
	Args:    requireSandboxArg,
	Aliases: common.GetAliases("preview-url"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient_cli.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		sandboxIdOrName := args[0]

		if previewUrlPort == 0 {
			return clierr.New(clierr.CategoryUsage, "port flag is required")
		}

		req := apiClient.SandboxAPI.GetSignedPortPreviewUrl(ctx, sandboxIdOrName, previewUrlPort).
			ExpiresInSeconds(previewUrlExpires)

		previewUrl, res, err := req.Execute()
		if err != nil {
			return apiclient_cli.HandleErrorResponse(res, err)
		}

		if common.FormatFlag != "" {
			common.NewFormatter(newPreviewUrlOutput(previewUrl, previewUrlExpires, sandboxIdOrName)).Print()
			return nil
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
	common.RegisterFormatFlag(PreviewUrlCmd)
}
