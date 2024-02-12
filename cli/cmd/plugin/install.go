// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_plugin

import (
	"context"
	"log"

	"github.com/daytonaio/daytona/cli/cmd/views/plugins/list_view"
	"github.com/daytonaio/daytona/cli/connection"
	"github.com/daytonaio/daytona/common/grpc/proto"
	"github.com/daytonaio/daytona/common/os"
	"github.com/daytonaio/daytona/plugins/plugin_manager"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/spf13/cobra"
)

var pluginInstallCmd = &cobra.Command{
	Use:     "install",
	Short:   "Install plugin",
	Args:    cobra.NoArgs,
	Aliases: []string{"i"},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		conn, err := connection.Get(nil)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		serverClient := proto.NewServerClient(conn)

		config, err := serverClient.GetConfig(ctx, &empty.Empty{})
		if err != nil {
			log.Fatal(err)
		}

		pluginsManifest, err := plugin_manager.GetPluginsManifest(config.PluginRegistryUrl)
		if err != nil {
			log.Fatal(err)
		}

		pluginList := convertToDTO(pluginsManifest.ProvisionerPlugins, list_view.PluginTypeProvisioner)
		pluginList = append(pluginList, convertToDTO(pluginsManifest.AgentServicePlugins, list_view.PluginTypeAgentService)...)

		pluginToInstall := list_view.GetPluginFromPrompt(pluginList, "Choose a Plugin to install")

		if pluginToInstall == nil {
			return
		}

		pluginsClient := proto.NewPluginsClient(conn)

		if pluginToInstall.Type == list_view.PluginTypeProvisioner {
			_, err = pluginsClient.InstallProvisionerPlugin(ctx, &proto.InstallPluginRequest{
				Name:         pluginToInstall.Name,
				DownloadUrls: convertToStringMap(pluginsManifest.ProvisionerPlugins[pluginToInstall.Name].Versions[pluginToInstall.Version].DownloadUrls),
			})
		} else if pluginToInstall.Type == list_view.PluginTypeAgentService {
			_, err = pluginsClient.InstallAgentServicePlugin(ctx, &proto.InstallPluginRequest{
				Name:         pluginToInstall.Name,
				DownloadUrls: convertToStringMap(pluginsManifest.AgentServicePlugins[pluginToInstall.Name].Versions[pluginToInstall.Version].DownloadUrls),
			})
		} else {
			log.Fatal("Unknown plugin type")
		}

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
