// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_workspace

import (
	"context"
	"dagent/client"
	select_prompt "dagent/cmd/views/workspace_select_prompt"
	workspace_proto "dagent/grpc/proto"
	"os"

	"github.com/golang/protobuf/ptypes/empty"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var startProjectFlag string

var StartCmd = &cobra.Command{
	Use:   "start [WORKSPACE_NAME]",
	Short: "Start the workspace",
	Args:  cobra.MaximumNArgs(1),
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

			workspaceName = select_prompt.GetWorkspaceNameFromPrompt(workspaceList.Workspaces, "start")
		} else {
			workspaceName = args[0]
		}

		wsName, wsMode := os.LookupEnv("DAYTONA_WS_NAME")
		if wsMode {
			workspaceName = wsName
		}

		startWorkspaceRequest := &workspace_proto.WorkspaceStartRequest{
			Name:    workspaceName,
			Project: startProjectFlag,
		}
		_, err = client.Start(ctx, startWorkspaceRequest)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	_, exists := os.LookupEnv("DAYTONA_WS_DIR")
	if exists {
		StartCmd.Use = "start"
		StartCmd.Args = cobra.ExactArgs(0)
	}

	StartCmd.PersistentFlags().StringVarP(&startProjectFlag, "project", "p", "", "Start the single project in the workspace (project name)")
}
