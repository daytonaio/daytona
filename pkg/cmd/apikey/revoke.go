// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikey

import (
	"context"
	"errors"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/apikeys"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/server/apikey"
)

var yesFlag bool

var revokeCmd = &cobra.Command{
	Use:     "revoke [NAME]",
	Short:   "Revoke an API key",
	Aliases: []string{"r", "rm", "delete"},
	Args:    cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			return err
		}

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		var selectedApiKey *apiclient.ApiKey

		apiKeyList, _, err := apiClient.ApiKeyAPI.ListClientApiKeys(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(nil, err)
		}

		if len(args) == 1 {
			for _, apiKey := range apiKeyList {
				if apiKey.Name == args[0] {
					selectedApiKey = &apiKey
					break
				}
			}
		} else {
			selectedApiKey, err = apikey.GetApiKeyFromPrompt(apiKeyList, "Select an API key to revoke", false)
			if err != nil {
				if common.IsCtrlCAbort(err) {
					return nil
				} else {
					return err
				}
			}
		}

		if selectedApiKey == nil {
			return errors.New("No API key selected")
		}

		if !yesFlag {
			title := fmt.Sprintf("Revoke API Key '%s'?", selectedApiKey.Name)
			description := fmt.Sprintf("Are you sure you want to revoke '%s'?", selectedApiKey.Name)
			if apikeys.EqualsKeyHashFromApi(activeProfile.Api.Key, selectedApiKey.KeyHash) {
				title = fmt.Sprintf("Warning! API Key '%s' is attached to your active profile", selectedApiKey.Name)
				description = fmt.Sprintf("Revoking '%s' will lock out your active profile from accessing the server.", selectedApiKey.Name)
			}

			form := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title(title).
						Description(description).
						Value(&yesFlag),
				),
			).WithTheme(views.GetCustomTheme())

			err := form.Run()
			if err != nil {
				return err
			}
		}

		if yesFlag {
			res, err := apiClient.ApiKeyAPI.RevokeApiKey(ctx, selectedApiKey.Name).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			views.RenderInfoMessage("API key revoked")
		} else {
			fmt.Println("Operation canceled.")
		}

		return nil
	},
}

func init() {
	revokeCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip confirmation prompt")
}
