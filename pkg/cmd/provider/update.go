// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"fmt"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	cmd_common "github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/views/provider"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/spf13/cobra"
)

var allFlag bool

var updateCmd = &cobra.Command{
	Use:     "update",
	Short:   "Update provider",
	Args:    cobra.NoArgs,
	Aliases: cmd_common.GetAliases("update"),
	RunE: func(cmd *cobra.Command, args []string) error {
		var selectedRunnerId string

		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			selectedRunner, err := cmd_common.GetRunnerFlow(apiClient, "Manage Providers")
			if err != nil {
				if common.IsCtrlCAbort(err) {
					return nil
				} else {
					return err
				}
			}

			if selectedRunner == nil {
				return nil
			}

			selectedRunnerId = selectedRunner.Id
		} else {
			selectedRunnerId = args[0]
		}

		providerList, res, err := apiClient.ProviderAPI.ListProviders(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		if len(providerList) == 0 {
			views_util.NotifyEmptyProviderList(true)
			return nil
		}

		if allFlag {
			for _, provider := range providerList {
				fmt.Printf("Updating provider %s\n", provider.Name)
				res, err := apiClient.ProviderAPI.UpdateProvider(context.Background(), selectedRunnerId, provider.Name).Execute()
				if err != nil {
					return apiclient_util.HandleErrorResponse(res, err)
				} else {
					fmt.Printf("Provider %s has been successfully updated\n", provider.Name)
				}
			}

			return nil
		}

		providerToUpdate, err := provider.GetProviderFromPrompt(provider.ProviderListToView(providerList), "Choose a Provider to Update", false)
		if err != nil {
			if common.IsCtrlCAbort(err) {
				return nil
			} else {
				return err
			}
		}
		if providerToUpdate == nil {
			return nil
		}

		res, err = apiClient.ProviderAPI.UpdateProvider(context.Background(), selectedRunnerId, providerToUpdate.Name).ProviderVersion(providerToUpdate.Version).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		fmt.Printf("Provider %s update is in progress\n", providerToUpdate.Name)
		return nil
	},
}

func init() {
	updateCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Update all providers")
}
