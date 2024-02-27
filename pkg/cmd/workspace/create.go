// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"errors"
	"os"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/types"
	provider_view "github.com/daytonaio/daytona/pkg/views/provider"
	view_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/create"
	"github.com/daytonaio/daytona/pkg/views/workspace/info"
	"github.com/daytonaio/daytona/pkg/views/workspace/initialize"

	tea "github.com/charmbracelet/bubbletea"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/daytonaio/daytona/cmd/daytona/config"
)

var repos []string

var CreateCmd = &cobra.Command{
	Use:   "create [WORKSPACE_NAME]",
	Short: "Create a workspace",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		var workspaceName string
		var provider string

		manual, _ := cmd.Flags().GetBool("manual")
		apiClient, err := server.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		serverConfig, res, err := apiClient.ServerAPI.GetConfig(ctx).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			log.Fatal(err)
		}

		if providerFlag != "" {
			provider = providerFlag
		} else if activeProfile.DefaultProvider == "" {

			providersList, res, err := apiClient.ProviderAPI.ListProviders(context.Background()).Execute()
			if err != nil {
				log.Fatal(apiclient.HandleErrorResponse(res, err))
			}

			if len(providersList) == 0 {
				log.Fatal(errors.New("no provider plugins found"))
			}

			defaultProvider := provider_view.GetProviderFromPrompt(providersList, "Provider not set. Choose a provider to use")

			provider = *defaultProvider.Name
			activeProfile.DefaultProvider = *defaultProvider.Name

			err = c.EditProfile(activeProfile)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			provider = activeProfile.DefaultProvider
		}

		if len(args) == 0 {
			var workspaceNames []string
			repos = []string{} // Ignore repo flags if prompting

			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
			if err != nil {
				log.Fatal(apiclient.HandleErrorResponse(res, err))
			}
			for _, workspaceInfo := range workspaceList {
				workspaceNames = append(workspaceNames, *workspaceInfo.Name)
			}

			view_util.RenderMainTitle("WORKSPACE CREATION")

			var gitProviderList []types.GitProvider
			for _, serverGitProvider := range serverConfig.GitProviders {
				var gitProvider types.GitProvider
				if serverGitProvider.Id != nil {
					gitProvider.Id = *serverGitProvider.Id
				}
				if serverGitProvider.Username != nil {
					gitProvider.Username = *serverGitProvider.Username
				}
				if serverGitProvider.Token != nil {
					gitProvider.Token = *serverGitProvider.Token
				}
				gitProviderList = append(gitProviderList, gitProvider)
			}

			workspaceName, repos, err = create.GetCreationDataFromPrompt(workspaceNames, gitProviderList, manual)
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

		_, res, err = apiClient.WorkspaceAPI.CreateWorkspace(ctx).Workspace(serverapiclient.CreateWorkspace{
			Name:         &workspaceName,
			Repositories: repos,
			Provider:     &provider,
		}).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		initViewModel := initialize.GetInitialModel()
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

		// started := false
		// for {
		// 	if started {
		// 		break
		// 	}
		// 	select {
		// 	case <-stream.Context().Done():
		// 		started = true
		// 	default:
		// 		// Recieve on the stream
		// 		res, err := stream.Recv()
		// 		if err != nil {
		// 			if err == io.EOF {
		// 				started = true
		// 				break
		// 			} else {
		// 				initViewProgram.Send(tea.Quit())
		// 				initViewProgram.ReleaseTerminal()
		// 				log.Fatal(err)
		// 				return
		// 			}
		// 		}
		// 		initViewProgram.Send(init_view.EventMsg{Event: res.Event, Payload: res.Payload})
		// 	}
		// }

		wsInfo, res, err := apiClient.WorkspaceAPI.GetWorkspaceInfo(ctx, workspaceName).Execute()
		if err != nil {
			initViewProgram.Send(tea.Quit())
			initViewProgram.ReleaseTerminal()
			log.Fatal(apiclient.HandleErrorResponse(res, err))
			return
		}
		initViewProgram.Send(initialize.ClearScreenMsg{})
		initViewProgram.Send(tea.Quit())
		initViewProgram.ReleaseTerminal()

		//	Show the info
		info.Render(wsInfo)
	},
}

var providerFlag string

func init() {
	CreateCmd.Flags().StringArrayVarP(&repos, "repo", "r", nil, "Set the repository url")
	CreateCmd.Flags().StringVar(&providerFlag, "provider", "", "Specify the provider (e.g. 'docker-provider')")
	CreateCmd.Flags().Bool("manual", false, "Manually enter the git repositories")
}
