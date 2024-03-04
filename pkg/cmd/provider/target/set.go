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

var targetSetCmd = &cobra.Command{
	Use:     "set",
	Short:   "Set provider target",
	Args:    cobra.NoArgs,
	Aliases: []string{"s"},
	Run: func(cmd *cobra.Command, args []string) {
		pluginList, err := server.GetProviderList()
		if err != nil {
			log.Fatal(err)
		}

		selectedProvider := provider.GetProviderFromPrompt(pluginList, "Choose a Provider")

		if selectedProvider == nil {
			return
		}

		selectedTarget, err := target.GetTargetFromPrompt(selectedProvider.Targets, true)
		if err != nil {
			log.Fatal(err)
		}

		client, err := server.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		targetManifest, res, err := client.ProviderAPI.GetTargetManifest(context.Background(), *selectedProvider.Name).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		if *selectedTarget.Name == target.NewTargetName {
			*selectedTarget.Name = ""
			err = target.NewTargetNameInput(selectedTarget.Name)
			if err != nil {
				log.Fatal(err)
			}
		}

		err = target.SetTargetForm(selectedTarget, *targetManifest)
		if err != nil {
			log.Fatal(err)
		}

		res, err = client.ProviderAPI.SetTarget(context.Background(), *selectedProvider.Name).Target(*selectedTarget).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}
	},
}
