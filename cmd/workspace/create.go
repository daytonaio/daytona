// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_workspace

import (
	"context"
	"io"
	"os"

	"github.com/daytonaio/daytona/client"
	workspace_proto "github.com/daytonaio/daytona/grpc/proto"
	"github.com/daytonaio/daytona/internal/util"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/golang/protobuf/ptypes/empty"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	init_view "github.com/daytonaio/daytona/cmd/views/init_workspace"
	views_util "github.com/daytonaio/daytona/cmd/views/util"
	wizard_view "github.com/daytonaio/daytona/cmd/views/workspace_create_wizard"
	info_view "github.com/daytonaio/daytona/cmd/views/workspace_info"
)

var repos []string

var CreateCmd = &cobra.Command{
	Use:   "create [WORKSPACE_NAME]",
	Short: "Create the workspace",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		var workspaceName string

		conn, err := client.GetConn(nil)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()
		client := workspace_proto.NewWorkspaceClient(conn)

		if len(args) == 0 {
			var workspaceNames []string
			ctx := context.Background()

			workspaceListResponse, err := client.List(ctx, &empty.Empty{})
			if err != nil {
				log.Fatal(err)
			}
			for _, workspaceInfo := range workspaceListResponse.Workspaces {
				workspaceNames = append(workspaceNames, workspaceInfo.Name)
			}

			repos = []string{} // Ignore repo flags if prompting

			views_util.RenderMainTitle("WORKSPACE CREATION")

			workspaceName, repos, err = wizard_view.GetCreationDataFromPrompt(workspaceNames)
			if err != nil {
				log.Fatal(err)
				return
			}
		} else {
			validatedWorkspaceName, err := util.GetValidatedWorkspaceName(args[0])
			if err != nil {
				log.Fatal(err)
				return
			}
			workspaceName = validatedWorkspaceName
		}

		if workspaceName == "" || len(repos) == 0 {
			return
		}

		ctx := context.Background()

		createRequest := &workspace_proto.CreateWorkspaceRequest{
			Name:         workspaceName,
			Repositories: repos,
		}

		stream, err := client.Create(ctx, createRequest)
		if err != nil {
			log.Fatal(err)
		}

		initViewModel := init_view.GetInitialModel()
		initViewProgram := tea.NewProgram(initViewModel)

		go func() {
			_, err := initViewProgram.Run()
			initViewProgram.ReleaseTerminal()
			if err != nil {
				log.Fatal(err)
				os.Exit(1)
			}
			os.Exit(0)
		}()

		started := false
		for {
			if started {
				break
			}
			select {
			case <-stream.Context().Done():
				started = true
			default:
				// Recieve on the stream
				res, err := stream.Recv()
				if err != nil {
					if err == io.EOF {
						started = true
						break
					} else {
						initViewProgram.Send(tea.Quit())
						initViewProgram.ReleaseTerminal()
						log.Fatal(err)
						return
					}
				}
				initViewProgram.Send(init_view.EventMsg{Event: res.Event, Payload: res.Payload})
			}
		}

		infoWorkspaceRequest := &workspace_proto.WorkspaceInfoRequest{
			Name: workspaceName,
		}
		response, err := client.Info(ctx, infoWorkspaceRequest)
		if err != nil {
			initViewProgram.Send(tea.Quit())
			initViewProgram.ReleaseTerminal()
			log.Fatal(err)
			return
		}
		initViewProgram.Send(init_view.ClearScreenMsg{})
		initViewProgram.Send(tea.Quit())
		initViewProgram.ReleaseTerminal()

		//	Show the info
		info_view.Render(response)
	},
}

func init() {
	CreateCmd.Flags().StringArrayVarP(&repos, "repo", "r", nil, "Set the repository url")
}
