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

	log "github.com/sirupsen/logrus"
)

var providerUninstallCmd = &cobra.Command{
	Use:     "uninstall",
	Short:   "Uninstall provider",
	Args:    cobra.NoArgs,
	Aliases: []string{"u"},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		providerList, res, err := apiClient.ProviderAPI.ListProviders(ctx).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		providerToUninstall, err := provider.GetProviderFromPrompt(provider.ProviderListToView(providerList), "Choose a Provider to Uninstall", false)
		if err != nil {
			if common.IsCtrlCAbort(err) {
				return
			} else {
				log.Fatal(err)
			}
		}

		if providerToUninstall == nil {
			return
		}

		res, err = apiClient.ProviderAPI.UninstallProvider(ctx, providerToUninstall.Name).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		views.RenderInfoMessageBold(fmt.Sprintf("Provider %s has been successfully uninstalled", providerToUninstall.Name))
	},
}
