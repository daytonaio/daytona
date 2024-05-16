// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/daytonaio/daytona/internal/cmd/tailscale"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	ssh_config "github.com/daytonaio/daytona/pkg/agent/ssh/config"
	workspace_util "github.com/daytonaio/daytona/pkg/cmd/workspace/util"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/target"
	"github.com/daytonaio/daytona/pkg/views/workspace/info"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/docker/docker/pkg/stringid"
	"github.com/gorilla/websocket"
	"tailscale.com/tsnet"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/daytonaio/daytona/cmd/daytona/config"
)

var CreateCmd = &cobra.Command{
	Use:   "create [REPOSITORY_URL]",
	Short: "Create a workspace",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		var projects []serverapiclient.CreateWorkspaceRequestProject
		var workspaceName string
		var existingWorkspaceNames []string

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

		profileData, res, err := apiClient.ProfileAPI.GetProfileData(ctx).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		if nameFlag != "" {
			workspaceName = nameFlag
		}

		workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}
		for _, workspaceInfo := range workspaceList {
			existingWorkspaceNames = append(existingWorkspaceNames, *workspaceInfo.Name)
		}

		if len(args) == 0 {
			err = processPrompting(apiClient, &workspaceName, &projects, existingWorkspaceNames, ctx)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			err = processCmdArguments(args, apiClient, &projects, ctx)
			if err != nil {
				log.Fatal(err)
			}

			if workspaceName == "" {
				workspaceName = workspace_util.GetSuggestedWorkspaceName(projects[0].Name, existingWorkspaceNames)
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
			projects[i].EnvVars = getEnvVariables(&projects[i], profileData)
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
		id := stringid.GenerateRandomID()
		id = stringid.TruncateID(id)

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
var nameFlag string
var targetNameFlag string
var manualFlag bool
var multiProjectFlag bool
var codeFlag bool

func init() {
	CreateCmd.Flags().StringVar(&nameFlag, "name", "", "Specify the workspace name")
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

func processPrompting(apiClient *serverapiclient.APIClient, workspaceName *string, projects *[]serverapiclient.CreateWorkspaceRequestProject, workspaceNames []string, ctx context.Context) error {
	gitProviders, res, err := apiClient.GitProviderAPI.ListGitProviders(ctx).Execute()
	if err != nil {
		return apiclient.HandleErrorResponse(res, err)
	}

	apiServerConfig, res, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
	if err != nil {
		return apiclient.HandleErrorResponse(res, err)
	}

	*workspaceName, *projects, err = workspace_util.GetCreationDataFromPrompt(workspace_util.CreateDataPromptConfig{
		ApiServerConfig:        apiServerConfig,
		ExistingWorkspaceNames: workspaceNames,
		UserGitProviders:       gitProviders,
		Manual:                 manualFlag,
		MultiProject:           multiProjectFlag,
		ApiClient:              apiClient,
	})
	if err != nil {
		return err
	}
	return nil
}

func processCmdArguments(args []string, apiClient *serverapiclient.APIClient, projects *[]serverapiclient.CreateWorkspaceRequestProject, ctx context.Context) error {
	repoUrl := args[0]

	repoUrl, err := util.GetValidatedUrl(repoUrl)
	if err != nil {
		return err
	}

	encodedURLParam := url.QueryEscape(repoUrl)
	repoResponse, res, err := apiClient.GitProviderAPI.GetGitContext(ctx, encodedURLParam).Execute()
	if err != nil {
		return apiclient.HandleErrorResponse(res, err)
	}

	project := &serverapiclient.CreateWorkspaceRequestProject{
		Name: *repoResponse.Name,
		Source: &serverapiclient.CreateWorkspaceRequestProjectSource{
			Repository: repoResponse,
		},
	}

	*projects = append(*projects, *project)

	return nil
}

func readWorkspaceLogs(activeProfile config.Profile, workspaceId string, projects []serverapiclient.CreateWorkspaceRequestProject, stopLogs *bool) {
	var wg sync.WaitGroup
	for _, project := range projects {
		wg.Add(1)
		go func(project serverapiclient.CreateWorkspaceRequestProject) {
			defer wg.Done()
			query := "follow=true"

			for {
				ws, res, err := server.GetWebsocketConn(fmt.Sprintf("/log/workspace/%s/%s", workspaceId, project.Name), &activeProfile, &query)
				// We want to retry getting the logs if it fails
				if err != nil {
					log.Trace(apiclient.HandleErrorResponse(res, err))
					time.Sleep(500 * time.Millisecond)
					continue
				}

				readLog(ws, stopLogs)

				ws.Close()
			}
		}(project)
	}

	query := "follow=true"

	for {
		ws, res, err := server.GetWebsocketConn(fmt.Sprintf("/log/workspace/%s", workspaceId), &activeProfile, &query)
		// We want to retry getting the logs if it fails
		if err != nil {
			log.Trace(apiclient.HandleErrorResponse(res, err))
			time.Sleep(500 * time.Millisecond)
			continue
		}

		readLog(ws, stopLogs)
		ws.Close()
		break
	}

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

		dialConn, err := tsConn.Dial(context.Background(), "tcp", fmt.Sprintf("%s:%d", workspace.GetProjectHostname(workspaceId, projectName), ssh_config.SSH_PORT))
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

func getEnvVariables(project *serverapiclient.CreateWorkspaceRequestProject, profileData *serverapiclient.ProfileData) *map[string]string {
	envVars := map[string]string{}

	if profileData.EnvVars != nil {
		for k, v := range *profileData.EnvVars {
			if strings.HasPrefix(v, "$") {
				env, ok := os.LookupEnv(v[1:])
				if ok {
					envVars[k] = env
				} else {
					log.Warnf("Environment variable %s not found", v[1:])
				}
			} else {
				envVars[k] = v
			}
		}
	}

	if project.EnvVars != nil {
		for k, v := range *project.EnvVars {
			if strings.HasPrefix(v, "$") {
				env, ok := os.LookupEnv(v[1:])
				if ok {
					envVars[k] = env
				} else {
					log.Warnf("Environment variable %s not found", v[1:])
				}
			} else {
				envVars[k] = v
			}
		}
	}

	return &envVars
}
