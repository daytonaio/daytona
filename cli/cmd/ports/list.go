// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_ports

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/cli/cmd/output"
	"github.com/daytonaio/daytona/cli/connection"
	"github.com/daytonaio/daytona/common/grpc/proto"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var listPortForwardsCmd = &cobra.Command{
	Use:   "list [WORKSPACE_NAME] [PROJECT_NAME]",
	Short: "List port forwards",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := connection.Get(nil)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		projectName := ""

		if len(args) == 2 {
			projectName = args[1]
		}

		if projectName == "" {
			err = listPortForwards(conn, args[0], projectName)
		} else {
			err = listProjectPortForwards(conn, args[0], projectName)
		}

		if err != nil {
			log.Fatal(err)
		}
	},
}

func listPortForwards(conn *grpc.ClientConn, workspaceName string, projectName string) error {
	ctx := context.Background()

	client := proto.NewPortsClient(conn)
	getPortForwardsRequest := &proto.GetPortForwardsRequest{
		WorkspaceId: workspaceName,
	}

	response, err := client.GetPortForwards(ctx, getPortForwardsRequest)
	if err != nil {
		return err
	}

	for project, portForward := range response.ProjectPortForwards {
		fmt.Printf("Project: %s\n", project)
		for _, port := range portForward.PortForwards {
			if port.HostPort != port.ContainerPort {
				fmt.Printf("\tPort: %d (host: %d)\n", port.ContainerPort, port.HostPort)
			} else {
				fmt.Printf("\tPort: %d\n", port.ContainerPort)
			}
		}
	}

	if output.FormatFlag != "" {
		output.Output = response
	}

	return nil
}

func listProjectPortForwards(conn *grpc.ClientConn, workspaceName string, projectName string) error {
	projectPortForwards, err := GetProjectPortForwards(conn, workspaceName, projectName)
	if err != nil {
		return err
	}

	for _, port := range projectPortForwards.PortForwards {
		if port.HostPort != port.ContainerPort {
			fmt.Printf("Port: %d (host: %d)\n", port.ContainerPort, port.HostPort)
		} else {
			fmt.Printf("Port: %d\n", port.ContainerPort)
		}
	}

	if output.FormatFlag != "" {
		output.Output = projectPortForwards
	}

	return nil
}

func GetProjectPortForwards(conn *grpc.ClientConn, workspaceName string, projectName string) (*proto.ProjectPortForwards, error) {
	ctx := context.Background()

	client := proto.NewPortsClient(conn)
	getPortForwardsRequest := &proto.GetProjectPortForwardsRequest{
		WorkspaceId: workspaceName,
		Project:     projectName,
	}

	response, err := client.GetProjectPortForwards(ctx, getPortForwardsRequest)
	if err != nil {
		return nil, err
	}

	return response, nil
}
