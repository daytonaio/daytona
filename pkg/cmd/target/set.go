// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	internal_util "github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/provider"
	"github.com/daytonaio/daytona/pkg/views/target"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var TargetSetCmd = &cobra.Command{
	Use:     "set",
	Short:   "Set provider target",
	Args:    cobra.NoArgs,
	Aliases: []string{"s", "add", "update", "register", "edit"},
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			log.Fatal(err)
		}

		pluginList, err := server.GetProviderList()
		if err != nil {
			log.Fatal(err)
		}

		selectedProvider, err := provider.GetProviderFromPrompt(pluginList, "Choose a provider", false)
		if err != nil {
			log.Fatal(err)
		}

		if selectedProvider == nil {
			return
		}

		targets, err := server.GetTargetList()
		if err != nil {
			log.Fatal(err)
		}

		filteredTargets := []serverapiclient.ProviderTarget{}
		for _, t := range targets {
			if *t.ProviderInfo.Name == *selectedProvider.Name {
				filteredTargets = append(filteredTargets, t)
			}
		}

		selectedTarget, err := target.GetTargetFromPrompt(filteredTargets, activeProfile.Name, true)
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
			err = target.NewTargetNameInput(selectedTarget.Name, internal_util.ArrayMap(targets, func(t serverapiclient.ProviderTarget) string {
				return *t.Name
			}))
			if err != nil {
				log.Fatal(err)
			}
		}

		err = target.SetTargetForm(selectedTarget, *targetManifest)
		if err != nil {
			log.Fatal(err)
		}

		selectedTarget.ProviderInfo = &serverapiclient.ProviderProviderInfo{
			Name:    selectedProvider.Name,
			Version: selectedProvider.Version,
		}

		res, err = client.TargetAPI.SetTarget(context.Background()).Target(*selectedTarget).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		views.RenderInfoMessage("Target set successfully")
	},
}
