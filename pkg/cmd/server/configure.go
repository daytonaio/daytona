// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"

	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	server_view "github.com/daytonaio/daytona/pkg/views/server"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure Daytona Server",
	Run: func(cmd *cobra.Command, args []string) {
		apiClient, err := apiclient.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		apiServerConfig, res, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		containerRegistries, res, err := apiClient.ContainerRegistryAPI.ListContainerRegistries(context.Background()).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		apiServerConfig, err = server_view.ConfigurationForm(apiServerConfig, containerRegistries)
		if err != nil {
			log.Fatal(err)
		}

		_, res, err = apiClient.ServerAPI.SetConfig(context.Background()).Config(*apiServerConfig).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		views.RenderContainerLayout(views.GetInfoMessage("Server configuration updated. You need to restart the server for the changes to take effect."))
	},
}
