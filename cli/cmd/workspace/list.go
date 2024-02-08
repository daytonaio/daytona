// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_workspace

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/cli/cmd/output"
	"github.com/daytonaio/daytona/cli/connection"
	workspace_proto "github.com/daytonaio/daytona/common/grpc/proto"

	"github.com/golang/protobuf/ptypes/empty"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List workspaces",
	Args:    cobra.ExactArgs(0),
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		conn, err := connection.Get(nil)
		if err != nil {
			log.Fatal(err)
		}

		defer conn.Close()

		client := workspace_proto.NewWorkspaceServiceClient(conn)

		response, err := client.List(ctx, &empty.Empty{})
		if err != nil {
			log.Fatal(err)
		}

		if output.FormatFlag != "" {
			output.Output = response
			return
		}

		for _, workspaceInfo := range response.Workspaces {
			fmt.Println(workspaceInfo.Name)
		}
	},
}
