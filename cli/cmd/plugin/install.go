// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_plugin

import (
	"log"

	"github.com/daytonaio/daytona/cli/api"
	"github.com/daytonaio/daytona/cli/cmd/views/plugins/list_view"
	"github.com/daytonaio/daytona/common/api_client"
	"github.com/daytonaio/daytona/common/os"
	"github.com/daytonaio/daytona/plugins/plugin_manager"
	"github.com/spf13/cobra"
)

var pluginInstallCmd = &cobra.Command{
	Use:     "install",
	Short:   "Install plugin",
	Args:    cobra.NoArgs,
	Aliases: []string{"i"},
	Run: func(cmd *cobra.Command, args []string) {
		apiClient := api.GetServerApiClient("http://localhost:3000", "")

		serverConfig, _, err := apiClient.ServerAPI.GetConfigExecute(api_client.ApiGetConfigRequest{})
		if err != nil {
			log.Fatal(err)
		}

		pluginsManifest, err := plugin_manager.GetPluginsManifest(*serverConfig.PluginRegistryUrl)
		if err != nil {
			log.Fatal(err)
		}

		pluginList := convertToDTO(pluginsManifest.ProvisionerPlugins, list_view.PluginTypeProvisioner)
		pluginList = append(pluginList, convertToDTO(pluginsManifest.AgentServicePlugins, list_view.PluginTypeAgentService)...)

		pluginToInstall := list_view.GetPluginFromPrompt(pluginList, "Choose a Plugin to install")

		if pluginToInstall == nil {
			return
		}

		if pluginToInstall.Type == list_view.PluginTypeProvisioner {
			downloadUrls := convertToStringMap(pluginsManifest.ProvisionerPlugins[pluginToInstall.Name].Versions[pluginToInstall.Version].DownloadUrls)
			_, err = apiClient.PluginAPI.InstallProvisionerPluginExecute(api_client.ApiInstallProvisionerPluginRequest{}.Plugin(api_client.InstallPluginRequest{
				Name:         &pluginToInstall.Name,
				DownloadUrls: &downloadUrls,
			}))
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

func convertToDTO(plugins map[string]plugin_manager.PluginManifest, pluginType list_view.PluginType) []list_view.PluginViewDTO {
	pluginList := []list_view.PluginViewDTO{}
	for pluginName, pluginManifest := range plugins {
		for version := range pluginManifest.Versions {
			pluginList = append(pluginList, list_view.PluginViewDTO{
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
