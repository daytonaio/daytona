// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"

	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	server_view "github.com/daytonaio/daytona/pkg/views/server"
	"github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure Daytona Server",
	RunE: func(cmd *cobra.Command, args []string) error {
		apiClient, err := apiclient.GetApiClient(nil)
		if err != nil {
			return err
		}

		apiServerConfig, res, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		apiServerConfig, err = server_view.ConfigurationForm(apiServerConfig)
		if err != nil {
			return err
		}

		_, res, err = apiClient.ServerAPI.SetConfig(context.Background()).Config(*apiServerConfig).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		views.RenderContainerLayout(views.GetInfoMessage("Server configuration updated. You need to restart the server for the changes to take effect."))
		return nil
	},
}
