// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/internal/util/apiclient"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/provider"
	"github.com/spf13/cobra"
)

var providerUninstallCmd = &cobra.Command{
	Use:     "uninstall",
	Short:   "Uninstall provider",
	Args:    cobra.NoArgs,
	Aliases: []string{"u"},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		providerList, res, err := apiClient.ProviderAPI.ListProviders(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		providerToUninstall, err := provider.GetProviderFromPrompt(provider.ProviderListToView(providerList), "Choose a Provider to Uninstall", false)
		if err != nil {
			if common.IsCtrlCAbort(err) {
				return nil
			} else {
				return err
			}
		}

		if providerToUninstall == nil {
			return nil
		}

		res, err = apiClient.ProviderAPI.UninstallProvider(ctx, providerToUninstall.Name).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		views.RenderInfoMessageBold(fmt.Sprintf("Provider %s has been successfully uninstalled", providerToUninstall.Name))
		return nil
	},
}
