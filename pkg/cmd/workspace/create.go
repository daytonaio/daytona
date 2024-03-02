// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/daytonaio/daytona/internal/tailscale"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/types"
	"github.com/daytonaio/daytona/pkg/views/target"
	view_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/create"
	"github.com/daytonaio/daytona/pkg/views/workspace/info"
	status "github.com/daytonaio/daytona/pkg/views/workspace/status"
	"github.com/gorilla/websocket"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/daytonaio/daytona/cmd/daytona/config"
)

var argRepos []string

var CreateCmd = &cobra.Command{
	Use:   "create [WORKSPACE_NAME]",
	Short: "Create a workspace",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		var repoUrls []string
		var repos []serverapiclient.Repository
		var workspaceName string

		manual, err := cmd.Flags().GetBool("manual")
		if err != nil {
			log.Fatal(err)
		}
		multiProjectFlag, err := cmd.Flags().GetBool("multi-project")
		if err != nil {
			log.Fatal(err)
		}

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

		view_util.RenderMainTitle("WORKSPACE CREATION")

		if len(args) == 0 {
			var workspaceNames []string

			if argRepos != nil {
				view_util.RenderInfoMessage("Error: workspace name argument is required for this command")
				cmd.Help()
			}

			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
			if err != nil {
				log.Fatal(apiclient.HandleErrorResponse(res, err))
			}
			for _, workspaceInfo := range workspaceList {
				workspaceNames = append(workspaceNames, *workspaceInfo.Name)
			}

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

			workspaceName, repos, err = create.GetCreationDataFromPrompt(workspaceNames, gitProviderList, manual, multiProjectFlag)
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
			if argRepos != nil {
				repoUrls = argRepos
			} else {
				view_util.RenderInfoMessage("Error: --repo flag is required for this command")
				cmd.Help()
			}

			for _, repoUrl := range repoUrls {
				repo, res, err := apiClient.ServerAPI.GetGitContext(ctx, repoUrl).Execute()
				if err != nil {
					log.Fatal(apiclient.HandleErrorResponse(res, err))
				}
				repos = append(repos, *repo)
			}
		}

		if workspaceName == "" || len(repos) == 0 {
			log.Fatal("workspace name and repository urls are required")
			return
		}

		statusProgram := tea.NewProgram(status.NewModel())

		started := false

		go func() {
			hostRegex := regexp.MustCompile(`https*://(.*)`)
			host := hostRegex.FindStringSubmatch(activeProfile.Api.Url)[1]
			wsURL := fmt.Sprintf("ws://%s/log/workspace/%s?follow=true", host, workspaceName)
			var ws *websocket.Conn
			var res *http.Response
			var err error

			ws, res, err = websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				log.Fatal(apiclient.HandleErrorResponse(res, err))
			}

			defer ws.Close()

			for {
				_, msg, err := ws.ReadMessage()
				if err != nil {
					// TODO: needs refactor
					ws.Close()
					ws, _, _ = websocket.DefaultDialer.Dial(wsURL, nil)
					continue
				}

				statusProgram.Send(status.ResultMsg{Line: string(msg)})
				if started {
					statusProgram.Send(status.ResultMsg{Line: "END_SIGNAL"})
					break
				}
			}
		}()

		go func() {
			if _, err := statusProgram.Run(); err != nil {
				fmt.Println("Error running status program:", err)
				statusProgram.Send(status.ClearScreenMsg{})
				statusProgram.Send(tea.Quit())
				statusProgram.ReleaseTerminal()
				os.Exit(1)
			}
		}()

		activeProfile, err = c.GetActiveProfile()
		if err != nil {
			log.Fatal(err)
		}

		target, err := getTarget(apiClient)
		if err != nil {
			log.Fatal(err)
		}

		tsConn, err := tailscale.GetConnection(&activeProfile)
		if err != nil {
			log.Fatal(err)
		}

		createdWorkspace, res, err := apiClient.WorkspaceAPI.CreateWorkspace(ctx).Workspace(serverapiclient.CreateWorkspace{
			Name:         &workspaceName,
			Repositories: repos,
			Target:       target.Name,
		}).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}
		started = true

		dialStartTime := time.Now()
		dialTimeout := 3 * time.Minute
		statusProgram.Send(status.ResultMsg{Line: "Establishing connection with the workspace"})
		for {
			if time.Since(dialStartTime) > dialTimeout {
				log.Fatal("Timeout: dialing timed out after 3 minutes")
			}

			dialConn, err := tsConn.Dial(context.Background(), "tcp", fmt.Sprintf("%s-%s:2222", workspaceName, *createdWorkspace.Projects[0].Name))
			if err == nil {
				defer dialConn.Close()
				break
			}

			time.Sleep(time.Second)
		}

		wsInfo, res, err := apiClient.WorkspaceAPI.GetWorkspace(ctx, workspaceName).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
			return
		}

		statusProgram.Send(status.ClearScreenMsg{})
		statusProgram.Send(tea.Quit())
		statusProgram.ReleaseTerminal()

		fmt.Println()
		info.Render(wsInfo)

		skipIdeFlag, _ := cmd.Flags().GetBool("skip-ide")
		if skipIdeFlag {
			return
		}

		ide := c.DefaultIdeId
		if ideFlag != "" {
			ide = ideFlag
		}

		view_util.RenderInfoMessageBold("Opening the workspace in your preferred IDE")
		openIDE(ide, activeProfile, workspaceName, *wsInfo.Projects[0].Name)
	},
}

var providerFlag string
var targetNameFlag string

func init() {
	CreateCmd.Flags().StringArrayVarP(&argRepos, "repo", "r", nil, "Set the repository url")
	CreateCmd.Flags().StringVar(&providerFlag, "provider", "", "Specify the provider (e.g. 'docker-provider')")
	CreateCmd.Flags().StringVarP(&ideFlag, "ide", "i", "", "Specify the IDE ('vscode' or 'browser')")
	CreateCmd.Flags().StringVarP(&targetNameFlag, "target", "t", "", "Specify the target (e.g. 'local')")
	CreateCmd.Flags().Bool("manual", false, "Manually enter the git repositories")
	CreateCmd.Flags().Bool("multi-project", false, "Workspace with multiple projects/repos")
	CreateCmd.Flags().Bool("skip-ide", false, "Don't open the IDE after workspace creation")
}

func getTarget(apiClient *serverapiclient.APIClient) (*serverapiclient.ProviderTarget, error) {
	targets, err := server.GetTargetList()
	if err != nil {
		return nil, err
	}

	var selectedTarget *serverapiclient.ProviderTarget = nil

	if targetNameFlag != "" {
		for _, t := range targets {
			if *t.Name == targetNameFlag {
				selectedTarget = &t
				break
			}
		}
	}

	if selectedTarget == nil {
		selectedTarget, err = target.GetTargetFromPrompt(targets, false)
		if err != nil {
			return nil, err
		}
	}

	return selectedTarget, nil
}
