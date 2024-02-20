// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_workspace

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/daytonaio/daytona/common/grpc/proto"
	"github.com/daytonaio/daytona/internal/util"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/golang/protobuf/ptypes/empty"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/daytonaio/daytona/cli/api"
	views_provisioner "github.com/daytonaio/daytona/cli/cmd/views/provisioner"
	views_util "github.com/daytonaio/daytona/cli/cmd/views/util"
	wizard_view "github.com/daytonaio/daytona/cli/cmd/views/workspace/creation_wizard"
	info_view "github.com/daytonaio/daytona/cli/cmd/views/workspace/info_view"
	init_view "github.com/daytonaio/daytona/cli/cmd/views/workspace/init_view"
	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/cli/connection"
)

var repos []string

var CreateCmd = &cobra.Command{
	Use:   "create [WORKSPACE_NAME]",
	Short: "Create a workspace",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		var workspaceName string
		var provisioner string

		conn, err := connection.GetGrpcConn(nil)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()
		client := proto.NewWorkspaceServiceClient(conn)

		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			log.Fatal(err)
		}

		if provisionerFlag != "" {
			provisioner = provisionerFlag
		} else if activeProfile.DefaultProvisioner == "" {
			apiClient := api.GetServerApiClient("http://localhost:3000", "")

			provisionerPluginList, _, err := apiClient.PluginAPI.ListProvisionerPlugins(context.Background()).Execute()
			if err != nil {
				log.Fatal(err)
			}

			if len(provisionerPluginList) == 0 {
				log.Fatal(errors.New("no provisioner plugins found"))
			}

			defaultProvisioner, err := views_provisioner.GetProvisionerFromPrompt(provisionerPluginList, "Provisioner not set. Choose a provisioner to use", nil)
			if err != nil {
				log.Fatal(err)
			}

			provisioner = *defaultProvisioner.Name
			activeProfile.DefaultProvisioner = *defaultProvisioner.Name

			err = c.EditProfile(activeProfile)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			provisioner = activeProfile.DefaultProvisioner
		}

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

		createRequest := &proto.CreateWorkspaceRequest{
			Name:         workspaceName,
			Repositories: repos,
			Provisioner:  provisioner,
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

		infoWorkspaceRequest := &proto.WorkspaceInfoRequest{
			Id: workspaceName,
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

var provisionerFlag string

func init() {
	CreateCmd.Flags().StringArrayVarP(&repos, "repo", "r", nil, "Set the repository url")
	CreateCmd.Flags().StringVar(&provisionerFlag, "provisioner", "", "Specify the provisioner (e.g. 'docker-provisioner')")
}
