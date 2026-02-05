// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"fmt"

	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/cmd/common"
	"github.com/daytonaio/daytona/cli/config"
	views_common "github.com/daytonaio/daytona/cli/views/common"
	"github.com/spf13/cobra"
)

var AdbCmd = &cobra.Command{
	Use:     "adb [SANDBOX_ID]",
	Short:   "Get ADB connection info for Android sandbox",
	Long:    "Returns ADB connection information and SSH tunnel command for connecting to an Android sandbox",
	Args:    cobra.ExactArgs(1),
	Aliases: common.GetAliases("adb"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient_cli.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		sandboxId := args[0]

		// Get sandbox info to verify it exists
		sb, res, err := apiClient.SandboxAPI.GetSandbox(ctx, sandboxId).Execute()
		if err != nil {
			return apiclient_cli.HandleErrorResponse(res, err)
		}

		// Get config for proxy domain
		cfg, err := config.GetConfig()
		if err != nil {
			return err
		}

		activeProfile, err := cfg.GetActiveProfile()
		if err != nil {
			return err
		}

		// Parse the API URL to get the base domain
		proxyDomain := "proxy.daytona.io"
		if activeProfile.Api.Url != "" {
			// Extract domain from API URL
			// In production, this would be derived from the API URL
			proxyDomain = "proxy.daytona.io"
		}

		// For development/local, use localhost
		if activeProfile.Api.Url == "http://localhost:3000" {
			proxyDomain = "localhost"
		}

		// Calculate ADB port based on instance number
		// Default base port is 6520, instance 1 = 6520, instance 2 = 6521, etc.
		adbPort := 6520 // Default for instance 1
		sshGatewayPort := 2220

		// Display connection info
		views_common.RenderInfoMessageBold("Android Sandbox ADB Connection")
		fmt.Println()
		fmt.Printf("  Sandbox ID:     %s\n", sandboxId)
		if sb.Name != "" {
			fmt.Printf("  Sandbox Name:   %s\n", sb.Name)
		}
		fmt.Printf("  ADB Port:       %d\n", adbPort)
		fmt.Println()

		views_common.RenderInfoMessage("To connect Android Studio, create an SSH tunnel:")
		fmt.Println()
		fmt.Printf("  ssh -L 5555:localhost:%d -p %d %s@%s\n", adbPort, sshGatewayPort, sandboxId, proxyDomain)
		fmt.Println()

		views_common.RenderInfoMessage("Then connect ADB:")
		fmt.Println()
		fmt.Printf("  adb connect localhost:5555\n")
		fmt.Println()

		views_common.RenderInfoMessage("Or in Android Studio:")
		fmt.Println()
		fmt.Printf("  Device Manager > Pair using Wi-Fi > Enter: localhost:5555\n")
		fmt.Println()

		return nil
	},
}

func init() {
	// No additional flags needed for now
}
