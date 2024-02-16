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
	"google.golang.org/protobuf/types/known/emptypb"
)

var pluginListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List installed plugins",
	Args:    cobra.NoArgs,
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		plugins, err := getPluginList()
		if err != nil {
			log.Fatal(err)
		}

		list_view.ListPlugins(plugins)
	},
}

func getPluginList() ([]list_view.PluginViewDTO, error) {
	ctx := context.Background()

	conn, err := connection.GetGrpcConn(nil)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := proto.NewPluginsClient(conn)

	plugins := []list_view.PluginViewDTO{}

	listProvisionerPluginsResponse, err := client.ListProvisionerPlugins(ctx, &emptypb.Empty{})
	if err != nil {
		log.Fatal(err)
	}

	for _, plugin := range listProvisionerPluginsResponse.Plugins {
		plugins = append(plugins, list_view.PluginViewDTO{
			Name:    plugin.Name,
			Version: plugin.Version,
			Type:    list_view.PluginTypeProvisioner,
		})
	}

	listAgentServicePluginsResponse, err := client.ListAgentServicePlugins(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	for _, plugin := range listAgentServicePluginsResponse.Plugins {
		plugins = append(plugins, list_view.PluginViewDTO{
			Name:    plugin.Name,
			Version: plugin.Version,
			Type:    list_view.PluginTypeAgentService,
		})
	}

	return plugins, nil
}
