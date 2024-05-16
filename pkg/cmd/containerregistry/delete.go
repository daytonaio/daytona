// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package containerregistry

import (
	"context"
	"net/url"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	containerregistry_view "github.com/daytonaio/daytona/pkg/views/containerregistry"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var containerRegistryDeleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"remove", "rm"},
	Short:   "Delete a container registry",
	Args:    cobra.RangeArgs(0, 2),
	Run: func(cmd *cobra.Command, args []string) {
		var registryDto *apiclient.ContainerRegistry
		var selectedServer string

		apiClient, err := server.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		if len(args) == 0 {
			c, err := config.GetConfig()
			if err != nil {
				log.Fatal(err)
			}

			activeProfile, err := c.GetActiveProfile()
			if err != nil {
				log.Fatal(err)
			}

			containerRegistries, res, err := apiClient.ContainerRegistryAPI.ListContainerRegistries(context.Background()).Execute()
			if err != nil {
				log.Fatal(apiclient_util.HandleErrorResponse(res, err))
			}

			if len(containerRegistries) == 0 {
				views.RenderInfoMessage("No container registries found")
				return
			}

			registryDto, err = containerregistry_view.GetRegistryFromPrompt(containerRegistries, activeProfile.Name, false)
			if err != nil {
				log.Fatal(err)
			}

			selectedServer = *registryDto.Server
		} else {
			selectedServer = args[0]
		}

		res, err := apiClient.ContainerRegistryAPI.RemoveContainerRegistry(context.Background(), url.QueryEscape(selectedServer)).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		views.RenderInfoMessage("Container registry deleted successfully")
	},
}
