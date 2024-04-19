// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/daytonaio/daytona/internal/tailscale"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	workspace_util "github.com/daytonaio/daytona/pkg/cmd/workspace/util"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views/target"
	view_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/info"
	status "github.com/daytonaio/daytona/pkg/views/workspace/status"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/google/uuid"
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
			processPrompting(cmd, apiClient, &workspaceName, &projects, ctx)
		} else {
			processCmdArguments(cmd, args, apiClient, &workspaceName, &projects, ctx)
		}

		view_util.RenderMainTitle("WORKSPACE CREATION")

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

		statusProgram := tea.NewProgram(status.NewModel())

		started := false
		id := uuid.NewString()

		go scanWorkspaceLogs(activeProfile, id, statusProgram, &started)

		go func() {
			if _, err := statusProgram.Run(); err != nil {
				fmt.Println("Error running status program:", err)
				statusProgram.Send(status.ClearScreenMsg{})
				statusProgram.Send(tea.Quit())
				err := statusProgram.ReleaseTerminal()
				if err != nil {
					log.Error(err)
				}
				os.Exit(1)
			}
		}()

		createdWorkspace, res, err := apiClient.WorkspaceAPI.CreateWorkspace(ctx).Workspace(serverapiclient.CreateWorkspaceRequest{
			Id:       &id,
			Name:     &workspaceName,
			Target:   target.Name,
			Projects: projects,
		}).Execute()
		if err != nil {
			cleanUpTerminal(statusProgram, apiclient.HandleErrorResponse(res, err))
		}

		dialStartTime := time.Now()
		dialTimeout := 3 * time.Minute
		// statusProgram.Send(status.Message{Line: "Establishing connection with the workspace"})

		waitForDial(tsConn, *createdWorkspace.Id, *createdWorkspace.Projects[0].Name, dialStartTime, dialTimeout, statusProgram)

		started = true

		cleanUpTerminal(statusProgram, nil)

		wsInfo, res, err := apiClient.WorkspaceAPI.GetWorkspace(ctx, workspaceName).Execute()
		if err != nil {
			cleanUpTerminal(statusProgram, apiclient.HandleErrorResponse(res, err))
			return
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
		info.Render(wsInfo, chosenIde.Name)

		codeFlag, _ := cmd.Flags().GetBool("code")
		if !codeFlag {
			view_util.RenderCreationInfoMessage("Run 'daytona code' when you're ready to start developing")
			return
		}

		view_util.RenderCreationInfoMessage("Opening the workspace in your preferred editor ...")

		err = openIDE(chosenIdeId, activeProfile, *createdWorkspace.Id, *wsInfo.Projects[0].Name)
		if err != nil {
			log.Fatal(err)
		}
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
	CreateCmd.Flags().BoolP("code", "c", false, "Open the workspace in the IDE after workspace creation")
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

func processPrompting(cmd *cobra.Command, apiClient *serverapiclient.APIClient, workspaceName *string, projects *[]serverapiclient.CreateWorkspaceRequestProject, ctx context.Context) {
	manualFlag, err := cmd.Flags().GetBool("manual")
	if err != nil {
		log.Fatal(err)
	}
	multiProjectFlag, err := cmd.Flags().GetBool("multi-project")
	if err != nil {
		log.Fatal(err)
	}

	gitProviders, res, err := apiClient.GitProviderAPI.ListGitProviders(ctx).Execute()
	if err != nil {
		log.Fatal(apiclient.HandleErrorResponse(res, err))
	}

	var workspaceNames []string

	if argRepos != nil {
		view_util.RenderInfoMessage("Error: workspace name argument is required for this command")
		err := cmd.Help()
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(1)
	}

	workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
	if err != nil {
		log.Fatal(apiclient.HandleErrorResponse(res, err))
	}
	for _, workspaceInfo := range workspaceList {
		workspaceNames = append(workspaceNames, *workspaceInfo.Name)
	}

	apiServerConfig, res, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
	if err != nil {
		log.Fatal(apiclient.HandleErrorResponse(res, err))
	}

	*workspaceName, *projects, err = workspace_util.GetCreationDataFromPrompt(apiServerConfig, workspaceNames, gitProviders, manualFlag, multiProjectFlag)
	if err != nil {
		log.Fatal(err)
		return
	}
}

func processCmdArguments(cmd *cobra.Command, args []string, apiClient *serverapiclient.APIClient, workspaceName *string, projects *[]serverapiclient.CreateWorkspaceRequestProject, ctx context.Context) {
	var repoUrls []string

	validatedWorkspaceName, err := util.GetValidatedWorkspaceName(args[0])
	if err != nil {
		log.Fatal(err)
		return
	}
	*workspaceName = validatedWorkspaceName
	if argRepos != nil {
		repoUrls = argRepos
	} else {
		view_util.RenderInfoMessage("Error: --repo flag is required for this command")
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
			log.Fatal(apiclient.HandleErrorResponse(res, err))
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
}

func scanWorkspaceLogs(activeProfile config.Profile, workspaceId string, statusProgram *tea.Program, started *bool) {
	time.Sleep(2 * time.Second)

	query := "follow=true"
	ws, res, err := server.GetWebsocketConn(fmt.Sprintf("/log/workspace/%s", workspaceId), &activeProfile, &query)
	if err != nil {
		cleanUpTerminal(statusProgram, apiclient.HandleErrorResponse(res, err))
	}

	defer ws.Close()

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			return
		}

		delimiter := byte('\r')
		messages := splitWithDelimiter(string(msg), delimiter)

		for _, message := range messages {
			line := fmt.Sprintf("%s %s", view_util.CheckmarkSymbol, message)
			fmt.Print(line)
		}
		if *started {
			fmt.Print("\nWorkspace creation complete.")
			statusProgram.Send(status.Message{Line: "END_SIGNAL"})
			break
		}
		if len(messages) != 0 {
			fmt.Print("\n")
		}
	}
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

func waitForDial(tsConn *tsnet.Server, workspaceId string, projectName string, dialStartTime time.Time, dialTimeout time.Duration, statusProgram *tea.Program) {
	for {
		if time.Since(dialStartTime) > dialTimeout {
			cleanUpTerminal(statusProgram, errors.New("timeout: dialing timed out after 3 minutes"))
		}

		dialConn, err := tsConn.Dial(context.Background(), "tcp", fmt.Sprintf("%s:2222", workspace.GetProjectHostname(workspaceId, projectName)))
		if err == nil {
			defer dialConn.Close()
			break
		}

		time.Sleep(time.Second)
	}
	cleanUpTerminal(statusProgram, nil)
}

func cleanUpTerminal(statusProgram *tea.Program, err error) {
	statusProgram.Send(status.ClearScreenMsg{})
	statusProgram.Send(tea.Quit())
	releaseError := statusProgram.ReleaseTerminal()
	if releaseError != nil {
		log.Error(releaseError)
	}
	if err != nil {
		log.Fatal(err)
	}
}
