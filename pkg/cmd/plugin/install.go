// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_plugin

import (
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/os"
	"github.com/daytonaio/daytona/pkg/plugin_manager"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views/plugins"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var pluginInstallCmd = &cobra.Command{
	Use:     "install",
	Short:   "Install plugin",
	Args:    cobra.NoArgs,
	Aliases: []string{"i"},
	Run: func(cmd *cobra.Command, args []string) {
		apiClient, err := server.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		serverConfig, res, err := apiClient.ServerAPI.GetConfigExecute(serverapiclient.ApiGetConfigRequest{})
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		pluginsManifest, err := plugin_manager.GetPluginsManifest(*serverConfig.PluginRegistryUrl)
		if err != nil {
			log.Fatal(err)
		}

		pluginList := convertToDTO(pluginsManifest.ProvisionerPlugins, plugins.PluginTypeProvisioner)
		pluginList = append(pluginList, convertToDTO(pluginsManifest.AgentServicePlugins, plugins.PluginTypeAgentService)...)

		pluginToInstall := plugins.GetPluginFromPrompt(pluginList, "Choose a Plugin to install")

		if pluginToInstall == nil {
			return
		}

		if pluginToInstall.Type == plugins.PluginTypeProvisioner {
			downloadUrls := convertToStringMap(pluginsManifest.ProvisionerPlugins[pluginToInstall.Name].Versions[pluginToInstall.Version].DownloadUrls)
			res, err = apiClient.PluginAPI.InstallProvisionerPluginExecute(serverapiclient.ApiInstallProvisionerPluginRequest{}.Plugin(serverapiclient.InstallPluginRequest{
				Name:         &pluginToInstall.Name,
				DownloadUrls: &downloadUrls,
			}))
			if err != nil {
				log.Fatal(apiclient.HandleErrorResponse(res, err))
			}
		}
		// else if pluginToInstall.Type == list_view.PluginTypeAgentService {
		// 	_, err = pluginsClient.InstallAgentServicePlugin(ctx, &proto.InstallPluginRequest{
		// 		Name:         pluginToInstall.Name,
		// 		DownloadUrls: convertToStringMap(pluginsManifest.AgentServicePlugins[pluginToInstall.Name].Versions[pluginToInstall.Version].DownloadUrls),
		// 	})
		// }
		// else {
		// 	log.Fatal("Unknown plugin type")
		// }

		if err != nil {
			log.Fatal(err)
		}
	},
}

func convertToDTO(pluginMap map[string]plugin_manager.PluginManifest, pluginType plugins.PluginType) []plugins.PluginViewDTO {
	pluginList := []plugins.PluginViewDTO{}
	for pluginName, pluginManifest := range pluginMap {
		for version := range pluginManifest.Versions {
			pluginList = append(pluginList, plugins.PluginViewDTO{
				Name:    pluginName,
				Version: version,
				Type:    pluginType,
			})
		}
	}

	return pluginList
}

func convertToStringMap(downloadUrls map[os.OperatingSystem]string) map[string]string {
	stringMap := map[string]string{}
	for os, url := range downloadUrls {
		stringMap[string(os)] = url
	}

	return stringMap
}
