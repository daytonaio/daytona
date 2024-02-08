// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_workspace

import (
	"context"
	"os"

	"github.com/daytonaio/daytona/client"
	workspace_proto "github.com/daytonaio/daytona/grpc/proto"
	"github.com/daytonaio/daytona/output"

	"github.com/golang/protobuf/ptypes/empty"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	view "github.com/daytonaio/daytona/cmd/views/workspace_info"
	select_prompt "github.com/daytonaio/daytona/cmd/views/workspace_select_prompt"
)

var InfoCmd = &cobra.Command{
	Use:     "info [WORKSPACE_NAME]",
	Short:   "Show workspace info",
	Aliases: []string{"view"},
	Args:    cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		var workspaceName string

		conn, err := client.GetConn(nil)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		client := workspace_proto.NewWorkspaceClient(conn)

		if len(args) == 0 {
			workspaceList, err := client.List(ctx, &empty.Empty{})
			if err != nil {
				log.Fatal(err)
			}

			workspaceName = select_prompt.GetWorkspaceNameFromPrompt(workspaceList.Workspaces, "view")
		} else {
			workspaceName = args[0]
		}

		wsName, wsMode := os.LookupEnv("DAYTONA_WS_NAME")
		if wsMode {
			workspaceName = wsName
		}

		workspaceInfoRequest := &workspace_proto.WorkspaceInfoRequest{
			Name: workspaceName,
		}
		response, err := client.Info(ctx, workspaceInfoRequest)
		if err != nil {
			log.Fatal(err)
		}

		view.Render(response)

		output.Output = response
	},
}

func init() {
	_, exists := os.LookupEnv("DAYTONA_WS_DIR")
	if exists {
		InfoCmd.Use = "info"
		InfoCmd.Args = cobra.ExactArgs(0)
	}
}
