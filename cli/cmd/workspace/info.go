// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_workspace

import (
	"context"
	"os"

	workspace_proto "github.com/daytonaio/daytona/common/grpc/proto"

	"github.com/golang/protobuf/ptypes/empty"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/daytonaio/daytona/cli/cmd/output"
	view "github.com/daytonaio/daytona/cli/cmd/views/workspace/info_view"
	select_prompt "github.com/daytonaio/daytona/cli/cmd/views/workspace/select_prompt"
	"github.com/daytonaio/daytona/cli/connection"
)

var InfoCmd = &cobra.Command{
	Use:     "info [WORKSPACE_NAME]",
	Short:   "Show workspace info",
	Aliases: []string{"view"},
	Args:    cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		var workspaceName string

		conn, err := connection.GetGrpcConn(nil)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		client := workspace_proto.NewWorkspaceServiceClient(conn)

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
			Id: workspaceName,
		}
		response, err := client.Info(ctx, workspaceInfoRequest)
		if err != nil {
			log.Fatal(err)
		}

		if output.FormatFlag != "" {
			output.Output = response
			return
		}

		view.Render(response)
	},
}

func init() {
	_, exists := os.LookupEnv("DAYTONA_WS_DIR")
	if exists {
		InfoCmd.Use = "info"
		InfoCmd.Args = cobra.ExactArgs(0)
	}
}
