// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_server

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/daytonaio/daytona/cli/api"
	view "github.com/daytonaio/daytona/cli/cmd/views/server/configuration_prompt"
	views_util "github.com/daytonaio/daytona/cli/cmd/views/util"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure Daytona Server",
	Run: func(cmd *cobra.Command, args []string) {
		apiClient, err := api.GetServerApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		apiServerConfig, _, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
		if err != nil {
			log.Fatal(err)
		}

		apiServerConfig = view.ConfigurationForm(apiServerConfig)

		_, _, err = apiClient.ServerAPI.SetConfig(context.Background()).Config(*apiServerConfig).Execute()
		if err != nil {
			log.Fatal(err)
		}

		views_util.RenderInfoMessage("Server configuration updated. You might need to restart the server for the changes to take effect.")
	},
}
