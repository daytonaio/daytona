// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_workspace

import (
	"context"
	"fmt"
	"os"

	"github.com/daytonaio/daytona/client"
	views_util "github.com/daytonaio/daytona/cmd/views/util"
	select_prompt "github.com/daytonaio/daytona/cmd/views/workspace_select_prompt"
	workspace_proto "github.com/daytonaio/daytona/grpc/proto"

	"github.com/golang/protobuf/ptypes/empty"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var stopProjectFlag string

var StopCmd = &cobra.Command{
	Use:   "stop [WORKSPACE_NAME]",
	Short: "Stop the workspace",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		var workspaceName string

		conn, err := client.GetConn(nil)
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

			workspaceName = select_prompt.GetWorkspaceNameFromPrompt(workspaceList.Workspaces, "stop")
		} else {
			workspaceName = args[0]
		}

		wsName, wsMode := os.LookupEnv("DAYTONA_WS_NAME")
		if wsMode {
			workspaceName = wsName
		}

		stopWorkspaceRequest := &workspace_proto.WorkspaceStopRequest{
			Id:      workspaceName,
			Project: stopProjectFlag,
		}
		_, err = client.Stop(ctx, stopWorkspaceRequest)
		if err != nil {
			log.Fatal(err)
		}

		views_util.RenderInfoMessage(fmt.Sprintf("Workspace %s successfully stopped", workspaceName))
	},
}

func init() {
	_, exists := os.LookupEnv("DAYTONA_WS_DIR")
	if exists {
		StopCmd.Use = "stop"
		StopCmd.Args = cobra.ExactArgs(0)
	}

	StopCmd.Flags().StringVarP(&stopProjectFlag, "project", "p", "", "Stop the single project in the workspace (project name)")
}
