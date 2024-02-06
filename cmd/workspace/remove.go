// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_workspace

import (
	"context"
	"dagent/client"
	"dagent/config"
	workspace_proto "dagent/grpc/proto"
	"os"

	select_prompt "dagent/cmd/views/workspace_select_prompt"

	"github.com/golang/protobuf/ptypes/empty"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var force bool

var DeleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete the workspace",
	Aliases: []string{"remove", "rm"},
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			log.Fatal(err)
		}

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

		removeWorkspaceRequest := &workspace_proto.WorkspaceRemoveRequest{
			Name:  workspaceName,
			Force: force,
		}
		_, err = client.Remove(ctx, removeWorkspaceRequest)
		if err != nil {
			log.Fatal(err)
		}

		config.RemoveWorkspaceSshEntries(activeProfile.Id, workspaceName)
	},
}

func init() {
	DeleteCmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "Force the workspace removal")
}
