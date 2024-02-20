// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_server

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/daytonaio/daytona/cli/api"
	view "github.com/daytonaio/daytona/cli/cmd/views/server/configuration_prompt"
	views_util "github.com/daytonaio/daytona/cli/cmd/views/util"
	"github.com/daytonaio/daytona/common/api_client"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure Daytona Server",
	Run: func(cmd *cobra.Command, args []string) {
		apiClient := api.GetServerApiClient("http://localhost:3000", "")

		apiServerConfig, _, err := apiClient.ServerAPI.GetConfigExecute(api_client.ApiGetConfigRequest{})
		if err != nil {
			log.Fatal(err)
		}

		serverConfig := api.ToServerConfig(apiServerConfig)

		view.ConfigurationForm(serverConfig)

		_, _, err = apiClient.ServerAPI.SetConfigExecute(api_client.ApiSetConfigRequest{}.Config(*api.FromServerConfig(serverConfig)))
		if err != nil {
			log.Fatal(err)
		}

		views_util.RenderInfoMessage("Server configuration updated. You might need to restart the server for the changes to take effect.")
	},
}
