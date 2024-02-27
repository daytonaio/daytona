// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_plugin

import (
	"context"
	"net/http"

	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/views/plugins"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var pluginUninstallCmd = &cobra.Command{
	Use:     "uninstall",
	Short:   "Uninstall plugin",
	Args:    cobra.NoArgs,
	Aliases: []string{"u"},
	Run: func(cmd *cobra.Command, args []string) {
		pluginList, err := getPluginList()
		if err != nil {
			log.Fatal(err)
		}

		pluginToUninstall := plugins.GetPluginFromPrompt(pluginList, "Choose a Plugin to uninstall")

		if pluginToUninstall == nil {
			return
		}

		apiClient, err := server.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}
		ctx := context.Background()

		var res *http.Response
		if pluginToUninstall.Type == plugins.PluginTypeProvisioner {
			res, err = apiClient.PluginAPI.UninstallProvisionerPlugin(ctx, pluginToUninstall.Name).Execute()
		} else if pluginToUninstall.Type == plugins.PluginTypeAgentService {
			res, err = apiClient.PluginAPI.UninstallAgentServicePlugin(ctx, pluginToUninstall.Name).Execute()
		} else {
			log.Fatal("Unknown plugin type")
		}

		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}
	},
}
