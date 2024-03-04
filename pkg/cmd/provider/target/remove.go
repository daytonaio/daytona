// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"

	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/views/provider"
	"github.com/daytonaio/daytona/pkg/views/provider/target"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var targetRemoveCmd = &cobra.Command{
	Use:     "remove",
	Short:   "Remove provider target",
	Args:    cobra.NoArgs,
	Aliases: []string{"rm"},
	Run: func(cmd *cobra.Command, args []string) {
		pluginList, err := server.GetProviderList()
		if err != nil {
			log.Fatal(err)
		}

		selectedProvider := provider.GetProviderFromPrompt(pluginList, "Choose a Provider")

		if selectedProvider == nil {
			return
		}

		selectedTarget, err := target.GetTargetFromPrompt(selectedProvider.Targets, false)
		if err != nil {
			log.Fatal(err)
		}

		client, err := server.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		res, err := client.ProviderAPI.RemoveTarget(context.Background(), *selectedProvider.Name, *selectedTarget.Name).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}
	},
}
