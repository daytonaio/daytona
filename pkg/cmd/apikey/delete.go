// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikey

import (
	"context"
	"errors"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	cmd_common "github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/apikey"
)

var yesFlag bool

var deleteCmd = &cobra.Command{
	Use:     "delete [NAME]",
	Short:   "Delete an API key",
	Args:    cobra.RangeArgs(0, 1),
	Aliases: cmd_common.GetAliases("delete"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		var selectedApiKey *apiclient.ApiKeyViewDTO

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
			return errors.New("no API key selected")
		}

		if !yesFlag {
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title(fmt.Sprintf("Revoke API Key '%s'?", selectedApiKey.Name)).
						Description(fmt.Sprintf("Are you sure you want to revoke '%s'?", selectedApiKey.Name)).
						Value(&yesFlag),
				),
			).WithTheme(views.GetCustomTheme())

			err := form.Run()
			if err != nil {
				return err
			}
		}

		if selectedApiKey.Current {
			return errors.New("cannot revoke current API key")
		}

		if yesFlag {
			res, err := apiClient.ApiKeyAPI.DeleteApiKey(ctx, selectedApiKey.Name).Execute()
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
	deleteCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip confirmation prompt")
}
