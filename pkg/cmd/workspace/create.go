// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/daytonaio/daytona/internal/tailscale"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	workspace_util "github.com/daytonaio/daytona/pkg/cmd/workspace/util"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/target"
	"github.com/daytonaio/daytona/pkg/views/workspace/info"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"tailscale.com/tsnet"

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
		var projects []serverapiclient.CreateWorkspaceRequestProject
		var workspaceName string

		apiClient, err := server.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			log.Fatal(err)
		}

		if len(args) == 0 {
			err = processPrompting(cmd, apiClient, &workspaceName, &projects, ctx)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			err = processCmdArguments(cmd, args, apiClient, &workspaceName, &projects, ctx)
			if err != nil {
				log.Fatal(err)
			}
		}

		if workspaceName == "" || len(projects) == 0 {
			log.Fatal("workspace name and repository urls are required")
			return
		}

		visited := make(map[string]bool)

		for i := range projects {
			if projects[i].Source == nil || projects[i].Source.Repository == nil || projects[i].Source.Repository.Url == nil {
				log.Fatal("Error: repository url is required")
			}
			if visited[*projects[i].Source.Repository.Url] {
				log.Fatalf("Error: duplicate repository url: %s", *projects[i].Source.Repository.Url)
			}
			visited[*projects[i].Source.Repository.Url] = true
		}

		target, err := getTarget(activeProfile.Name)
		if err != nil {
			log.Fatal(err)
		}

		activeProfile, err = c.GetActiveProfile()
		if err != nil {
			log.Fatal(err)
		}

		tsConn, err := tailscale.GetConnection(&activeProfile)
		if err != nil {
			log.Fatal(err)
		}

		stopLogs := false
		id := uuid.NewString()

		go readWorkspaceLogs(activeProfile, id, projects, &stopLogs)

		createdWorkspace, res, err := apiClient.WorkspaceAPI.CreateWorkspace(ctx).Workspace(serverapiclient.CreateWorkspaceRequest{
			Id:       &id,
			Name:     &workspaceName,
			Target:   target.Name,
			Projects: projects,
		}).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		dialStartTime := time.Now()
		dialTimeout := 3 * time.Minute

		err = waitForDial(tsConn, *createdWorkspace.Id, *createdWorkspace.Projects[0].Name, dialStartTime, dialTimeout)
		if err != nil {
			log.Fatal(err)
		}

		stopLogs = true

		wsInfo, res, err := apiClient.WorkspaceAPI.GetWorkspace(ctx, workspaceName).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		chosenIdeId := c.DefaultIdeId
		if ideFlag != "" {
			chosenIdeId = ideFlag
		}

		ideList := config.GetIdeList()
		var chosenIde config.Ide

		for _, ide := range ideList {
			if ide.Id == chosenIdeId {
				chosenIde = ide
			}
		}

		fmt.Println()
		info.Render(wsInfo, chosenIde.Name, false)

		if !codeFlag {
			views.RenderCreationInfoMessage("Run 'daytona code' when you're ready to start developing")
			return
		}

		views.RenderCreationInfoMessage("Opening the workspace in your preferred editor ...")

		err = openIDE(chosenIdeId, activeProfile, *createdWorkspace.Id, *wsInfo.Projects[0].Name)
		if err != nil {
			log.Fatal(err)
		}
	},
}

var providerFlag string
var targetNameFlag string
var manualFlag bool
var multiProjectFlag bool
var codeFlag bool

func init() {
	CreateCmd.Flags().StringArrayVarP(&argRepos, "repo", "r", nil, "Set the repository url")
	CreateCmd.Flags().StringVar(&providerFlag, "provider", "", "Specify the provider (e.g. 'docker-provider')")
	CreateCmd.Flags().StringVarP(&ideFlag, "ide", "i", "", "Specify the IDE ('vscode' or 'browser')")
	CreateCmd.Flags().StringVarP(&targetNameFlag, "target", "t", "", "Specify the target (e.g. 'local')")
	CreateCmd.Flags().BoolVar(&manualFlag, "manual", false, "Manually enter the git repositories")
	CreateCmd.Flags().BoolVar(&multiProjectFlag, "multi-project", false, "Workspace with multiple projects/repos")
	CreateCmd.Flags().BoolVarP(&codeFlag, "code", "c", false, "Open the workspace in the IDE after workspace creation")
}

func getTarget(activeProfileName string) (*serverapiclient.ProviderTarget, error) {
	targets, err := server.GetTargetList()
	if err != nil {
		return nil, err
	}

	if targetNameFlag != "" {
		for _, t := range targets {
			if *t.Name == targetNameFlag {
				return &t, nil
			}
		}
		return nil, fmt.Errorf("target '%s' not found", targetNameFlag)
	}

	if len(targets) == 1 {
		return &targets[0], nil
	}

	return target.GetTargetFromPrompt(targets, activeProfileName, false)
}

func processPrompting(cmd *cobra.Command, apiClient *serverapiclient.APIClient, workspaceName *string, projects *[]serverapiclient.CreateWorkspaceRequestProject, ctx context.Context) error {
	gitProviders, res, err := apiClient.GitProviderAPI.ListGitProviders(ctx).Execute()
	if err != nil {
		return apiclient.HandleErrorResponse(res, err)
	}

	var workspaceNames []string

	if argRepos != nil {
		views.RenderInfoMessage("Error: workspace name argument is required for this command")
		err := cmd.Help()
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(1)
	}

	workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
	if err != nil {
		return apiclient.HandleErrorResponse(res, err)
	}
	for _, workspaceInfo := range workspaceList {
		workspaceNames = append(workspaceNames, *workspaceInfo.Name)
	}

	apiServerConfig, res, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
	if err != nil {
		return apiclient.HandleErrorResponse(res, err)
	}

	*workspaceName, *projects, err = workspace_util.GetCreationDataFromPrompt(apiServerConfig, workspaceNames, gitProviders, manualFlag, multiProjectFlag)
	if err != nil {
		return err
	}
	return nil
}

func processCmdArguments(cmd *cobra.Command, args []string, apiClient *serverapiclient.APIClient, workspaceName *string, projects *[]serverapiclient.CreateWorkspaceRequestProject, ctx context.Context) error {
	var repoUrls []string

	validatedWorkspaceName, err := util.GetValidatedWorkspaceName(args[0])
	if err != nil {
		return err
	}
	*workspaceName = validatedWorkspaceName
	if argRepos != nil {
		repoUrls = argRepos
	} else {
		views.RenderInfoMessage("Error: --repo flag is required for this command")
		err := cmd.Help()
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(1)
	}

	for _, repoUrl := range repoUrls {
		encodedURLParam := url.QueryEscape(repoUrl)
		repoResponse, res, err := apiClient.GitProviderAPI.GetGitContext(ctx, encodedURLParam).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		projectName := workspace_util.GetProjectNameFromRepo(repoUrl)

		project := &serverapiclient.CreateWorkspaceRequestProject{
			Name: projectName,
			Source: &serverapiclient.CreateWorkspaceRequestProjectSource{
				Repository: &serverapiclient.GitRepository{Url: repoResponse.Url},
			},
		}

		*projects = append(*projects, *project)
	}
	return nil
}

func readWorkspaceLogs(activeProfile config.Profile, workspaceId string, projects []serverapiclient.CreateWorkspaceRequestProject, stopLogs *bool) {
	time.Sleep(2 * time.Second)

	query := "follow=true"
	ws, res, err := server.GetWebsocketConn(fmt.Sprintf("/log/workspace/%s", workspaceId), &activeProfile, &query)
	if err != nil {
		log.Fatal(apiclient.HandleErrorResponse(res, err))
	}

	defer ws.Close()

	var wg sync.WaitGroup
	for _, project := range projects {
		wg.Add(1)
		go func(project serverapiclient.CreateWorkspaceRequestProject) {
			defer wg.Done()
			query := "follow=true"
			ws, res, err := server.GetWebsocketConn(fmt.Sprintf("/log/workspace/%s/%s", workspaceId, project.Name), &activeProfile, &query)
			if err != nil {
				log.Fatal(apiclient.HandleErrorResponse(res, err))
			}

			defer ws.Close()

			readLog(ws, stopLogs)
		}(project)
	}

	readLog(ws, stopLogs)
	wg.Wait()
}

func splitWithDelimiter(s string, delimiter byte) []string {
	var parts []string
	var buffer []byte

	for i := 0; i < len(s); i++ {
		if s[i] == delimiter {
			parts = append(parts, string(buffer))
			buffer = nil
			parts = append(parts, string(delimiter))
		} else {
			buffer = append(buffer, s[i])
		}
	}

	// Add the remaining characters in the buffer
	if len(buffer) > 0 {
		parts = append(parts, string(buffer))
	}

	return parts
}

func waitForDial(tsConn *tsnet.Server, workspaceId string, projectName string, dialStartTime time.Time, dialTimeout time.Duration) error {
	for {
		if time.Since(dialStartTime) > dialTimeout {
			return errors.New("timeout: dialing timed out after 3 minutes")
		}

		dialConn, err := tsConn.Dial(context.Background(), "tcp", fmt.Sprintf("%s:2222", workspace.GetProjectHostname(workspaceId, projectName)))
		if err == nil {
			defer dialConn.Close()
			break
		}

		time.Sleep(time.Second)
	}
	return nil
}

func readLog(ws *websocket.Conn, stopLogs *bool) {
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			return
		}

		delimiter := byte('\r')
		messages := splitWithDelimiter(string(msg), delimiter)

		for _, message := range messages {
			fmt.Print(message)
		}
		if len(messages) != 0 {
			fmt.Print("\n")
		}

		if *stopLogs {
			return
		}
	}
}
