// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_plugin

import (
	"context"
	"log"

	"github.com/daytonaio/daytona/cli/cmd/views/plugins/list_view"
	"github.com/daytonaio/daytona/cli/connection"
	"github.com/daytonaio/daytona/common/grpc/proto"
	"github.com/spf13/cobra"
)

var pluginUninstallCmd = &cobra.Command{
	Use:     "uninstall",
	Short:   "Uninstall plugin",
	Args:    cobra.NoArgs,
	Aliases: []string{"u"},
	Run: func(cmd *cobra.Command, args []string) {
		plugins, err := getPluginList()
		if err != nil {
			log.Fatal(err)
		}

		pluginToUninstall := list_view.GetPluginFromPrompt(plugins, "Choose a Plugin to uninstall")

		if pluginToUninstall == nil {
			return
		}

		ctx := context.Background()

		conn, err := connection.Get(nil)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		client := proto.NewPluginsClient(conn)

		if pluginToUninstall.Type == list_view.PluginTypeProvisioner {
			_, err = client.UninstallProvisionerPlugin(ctx, &proto.UninstallPluginRequest{
				Name: pluginToUninstall.Name,
			})
		} else if pluginToUninstall.Type == list_view.PluginTypeAgentService {
			_, err = client.UninstallAgentServicePlugin(ctx, &proto.UninstallPluginRequest{
				Name: pluginToUninstall.Name,
			})
		} else {
			log.Fatal("Unknown plugin type")
		}

		if err != nil {
			log.Fatal(err)
		}
	},
}
