// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"

	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views/provider"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var providerListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List installed providers",
	Args:    cobra.NoArgs,
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		pluginList, err := getProviderList()
		if err != nil {
			log.Fatal(err)
		}

		provider.List(pluginList)
	},
}

func getProviderList() ([]serverapiclient.Provider, error) {
	apiClient, err := server.GetApiClient(nil)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	providersList, res, err := apiClient.ProviderAPI.ListProviders(ctx).Execute()
	if err != nil {
		log.Fatal(apiclient.HandleErrorResponse(res, err))
	}

	return providersList, nil
}
