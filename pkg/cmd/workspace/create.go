// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"github.com/daytonaio/daytona/internal/cmd/tailscale"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	ssh_config "github.com/daytonaio/daytona/pkg/agent/ssh/config"
	"github.com/daytonaio/daytona/pkg/apiclient"
	workspace_util "github.com/daytonaio/daytona/pkg/cmd/workspace/util"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/views"
	logs_view "github.com/daytonaio/daytona/pkg/views/logs"
	"github.com/daytonaio/daytona/pkg/views/target"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/create"
	"github.com/daytonaio/daytona/pkg/views/workspace/info"
	"github.com/daytonaio/daytona/pkg/workspace/project"
	"github.com/docker/docker/pkg/stringid"
	"tailscale.com/tsnet"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/daytonaio/daytona/cmd/daytona/config"
)

var CreateCmd = &cobra.Command{
	Use:     "create [REPOSITORY_URL]",
	Short:   "Create a workspace",
	Args:    cobra.RangeArgs(0, 1),
	GroupID: util.WORKSPACE_GROUP,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		var projects []apiclient.CreateProjectDTO
		var workspaceName string
		var existingWorkspaceNames []string
		var existingProjectConfigName *string

		apiClient, err := apiclient_util.GetApiClient(nil)
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
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		if nameFlag != "" {
			workspaceName = nameFlag
		}

		workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}
		for _, workspaceInfo := range workspaceList {
			existingWorkspaceNames = append(existingWorkspaceNames, workspaceInfo.Name)
		}

		if len(args) == 0 {
			err = processPrompting(apiClient, &workspaceName, &projects, existingWorkspaceNames, ctx)
			if err != nil {
				if common.IsCtrlCAbort(err) {
					return
				} else {
					log.Fatal(err)
				}
			}
		} else {
			existingProjectConfigName, err = processCmdArgument(args[0], apiClient, &projects, ctx)
			if err != nil {
				log.Fatal(err)
			}

			initialSuggestion := projects[0].Name

			if workspaceName == "" {
				workspaceName = workspace_util.GetSuggestedName(initialSuggestion, existingWorkspaceNames)
			}
		}

		if workspaceName == "" || len(projects) == 0 {
			log.Fatal("workspace name and repository urls are required")
			return
		}

		projectNames := []string{}
		for i := range projects {
			if profileData != nil && profileData.EnvVars != nil {
				projects[i].EnvVars = util.MergeEnvVars(profileData.EnvVars, projects[i].EnvVars)
			} else {
				projects[i].EnvVars = util.MergeEnvVars(projects[i].EnvVars)
			}
			projectNames = append(projectNames, projects[i].Name)
		}

		logs_view.CalculateLongestPrefixLength(projectNames)

		logs_view.DisplayLogEntry(logs.LogEntry{
			Msg: "Request submitted\n",
		}, logs_view.STATIC_INDEX)

		if existingProjectConfigName != nil {
			logs_view.DisplayLogEntry(logs.LogEntry{
				ProjectName: existingProjectConfigName,
				Msg:         fmt.Sprintf("Using detected project config '%s'\n", *existingProjectConfigName),
			}, logs_view.FIRST_PROJECT_INDEX)
		}

		targetList, res, err := apiClient.TargetAPI.ListTargets(ctx).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		target, err := getTarget(targetList, activeProfile.Name)
		if err != nil {
			log.Fatal(err)
		}

		activeProfile, err = c.GetActiveProfile()
		if err != nil {
			log.Fatal(err)
		}

		var tsConn *tsnet.Server
		if target.Name != "local" {
			tsConn, err = tailscale.GetConnection(&activeProfile)
			if err != nil {
				log.Fatal(err)
			}
		}

		id := stringid.GenerateRandomID()
		id = stringid.TruncateID(id)

		logsContext, stopLogs := context.WithCancel(context.Background())
		go apiclient_util.ReadWorkspaceLogs(logsContext, activeProfile, id, projectNames)

		createdWorkspace, res, err := apiClient.WorkspaceAPI.CreateWorkspace(ctx).Workspace(apiclient.CreateWorkspaceDTO{
			Id:       id,
			Name:     workspaceName,
			Target:   target.Name,
			Projects: projects,
		}).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		err = waitForDial(createdWorkspace, &activeProfile, tsConn)
		if err != nil {
			log.Fatal(err)
		}

		stopLogs()

		// Make sure terminal cursor is reset
		fmt.Print("\033[?25h")

		wsInfo, res, err := apiClient.WorkspaceAPI.GetWorkspace(ctx, workspaceName).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
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

		views.RenderCreationInfoMessage(fmt.Sprintf("Opening the workspace in %s ...", chosenIde.Name))

		projectName := wsInfo.Projects[0].Name
		providerMetadata, err := workspace_util.GetProjectProviderMetadata(wsInfo, projectName)
		if err != nil {
			log.Fatal(err)
		}

		err = openIDE(chosenIdeId, activeProfile, createdWorkspace.Id, projectName, providerMetadata)
		if err != nil {
			log.Fatal(err)
		}
	},
}

var providerFlag string
var nameFlag string
var targetNameFlag string
var codeFlag bool
var blankFlag bool
var multiProjectFlag bool

var projectConfigurationFlags = workspace_util.ProjectConfigurationFlags{
	Builder:          new(views_util.BuildChoice),
	CustomImage:      new(string),
	CustomImageUser:  new(string),
	Branch:           new(string),
	DevcontainerPath: new(string),
	EnvVars:          new([]string),
	Manual:           new(bool),
}

func init() {
	ideList := config.GetIdeList()
	ids := make([]string, len(ideList))
	for i, ide := range ideList {
		ids[i] = ide.Id
	}
	ideListStr := strings.Join(ids, ", ")

	CreateCmd.Flags().StringVar(&nameFlag, "name", "", "Specify the workspace name")
	CreateCmd.Flags().StringVar(&providerFlag, "provider", "", "Specify the provider (e.g. 'docker-provider')")
	CreateCmd.Flags().StringVarP(&ideFlag, "ide", "i", "", fmt.Sprintf("Specify the IDE (%s)", ideListStr))
	CreateCmd.Flags().StringVarP(&targetNameFlag, "target", "t", "", "Specify the target (e.g. 'local')")
	CreateCmd.Flags().BoolVar(&blankFlag, "blank", false, "Create a blank project without using existing configurations")
	CreateCmd.Flags().BoolVarP(&codeFlag, "code", "c", false, "Open the workspace in the IDE after workspace creation")
	CreateCmd.Flags().BoolVar(&multiProjectFlag, "multi-project", false, "Workspace with multiple projects/repos")

	workspace_util.AddProjectConfigurationFlags(CreateCmd, projectConfigurationFlags, true)
}

func getTarget(targetList []apiclient.ProviderTarget, activeProfileName string) (*apiclient.ProviderTarget, error) {
	if targetNameFlag != "" {
		for _, t := range targetList {
			if t.Name == targetNameFlag {
				return &t, nil
			}
		}
		return nil, fmt.Errorf("target '%s' not found", targetNameFlag)
	}

	if len(targetList) == 1 {
		return &targetList[0], nil
	}

	return target.GetTargetFromPrompt(targetList, activeProfileName, false)
}

func processPrompting(apiClient *apiclient.APIClient, workspaceName *string, projects *[]apiclient.CreateProjectDTO, workspaceNames []string, ctx context.Context) error {
	if workspace_util.CheckAnyProjectConfigurationFlagSet(projectConfigurationFlags) {
		return fmt.Errorf("please provide the repository URL in order to set up custom project details through the CLI")
	}

	gitProviders, res, err := apiClient.GitProviderAPI.ListGitProviders(ctx).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}

	projectConfigs, res, err := apiClient.ProjectConfigAPI.ListProjectConfigs(ctx).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}

	apiServerConfig, res, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}

	projectDefaults := &views_util.ProjectConfigDefaults{
		BuildChoice:          views_util.AUTOMATIC,
		Image:                &apiServerConfig.DefaultProjectImage,
		ImageUser:            &apiServerConfig.DefaultProjectUser,
		DevcontainerFilePath: create.DEVCONTAINER_FILEPATH,
	}

	*projects, err = workspace_util.GetProjectsCreationDataFromPrompt(workspace_util.ProjectsDataPromptConfig{
		UserGitProviders: gitProviders,
		ProjectConfigs:   projectConfigs,
		Manual:           *projectConfigurationFlags.Manual,
		MultiProject:     multiProjectFlag,
		BlankProject:     blankFlag,
		ApiClient:        apiClient,
		Defaults:         projectDefaults,
	})

	if err != nil {
		return err
	}

	initialSuggestion := (*projects)[0].Name

	suggestedName := workspace_util.GetSuggestedName(initialSuggestion, workspaceNames)

	submissionFormConfig := create.SubmissionFormConfig{
		ChosenName:    workspaceName,
		SuggestedName: suggestedName,
		ExistingNames: workspaceNames,
		ProjectList:   projects,
		NameLabel:     "Workspace",
		Defaults:      projectDefaults,
	}

	err = create.RunSubmissionForm(submissionFormConfig)
	if err != nil {
		return err
	}

	return nil
}

func processCmdArgument(argument string, apiClient *apiclient.APIClient, projects *[]apiclient.CreateProjectDTO, ctx context.Context) (*string, error) {
	if *projectConfigurationFlags.Builder != "" && *projectConfigurationFlags.Builder != views_util.DEVCONTAINER && *projectConfigurationFlags.DevcontainerPath != "" {
		return nil, fmt.Errorf("can't set devcontainer file path if builder is not set to %s", views_util.DEVCONTAINER)
	}

	var projectConfig *apiclient.ProjectConfig

	repoUrl, err := util.GetValidatedUrl(argument)
	if err == nil {
		// The argument is a Git URL
		return processGitURL(repoUrl, apiClient, projects, ctx)
	}

	// The argument is not a Git URL - try getting the project config
	projectConfig, _, err = apiClient.ProjectConfigAPI.GetProjectConfig(ctx, argument).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to parse the URL or fetch the project config for '%s'", argument)
	}

	return workspace_util.AddProjectFromConfig(projectConfig, apiClient, projects, *projectConfigurationFlags.Branch)
}

func processGitURL(repoUrl string, apiClient *apiclient.APIClient, projects *[]apiclient.CreateProjectDTO, ctx context.Context) (*string, error) {
	encodedURLParam := url.QueryEscape(repoUrl)

	if !blankFlag {
		projectConfig, res, err := apiClient.ProjectConfigAPI.GetDefaultProjectConfig(ctx, encodedURLParam).Execute()
		if err == nil {
			return workspace_util.AddProjectFromConfig(projectConfig, apiClient, projects, *projectConfigurationFlags.Branch)
		}

		if res.StatusCode != http.StatusNotFound {
			return nil, apiclient_util.HandleErrorResponse(res, err)
		}
	}

	var branch *string
	if *projectConfigurationFlags.Branch != "" {
		branch = projectConfigurationFlags.Branch
	}

	repo, res, err := apiClient.GitProviderAPI.GetGitContext(ctx).Repository(apiclient.GetRepositoryContext{
		Url:    repoUrl,
		Branch: branch,
	}).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	projectName, err := workspace_util.GetSanitizedProjectName(repo.Name)
	if err != nil {
		return nil, err
	}

	project, err := workspace_util.GetCreateProjectDtoFromFlags(projectConfigurationFlags)
	if err != nil {
		return nil, err
	}

	project.Name = projectName
	project.Source = apiclient.CreateProjectSourceDTO{
		Repository: *repo,
	}

	*projects = append(*projects, *project)

	return nil, nil
}

func waitForDial(workspace *apiclient.Workspace, activeProfile *config.Profile, tsConn *tsnet.Server) error {
	if workspace.Target == "local" {
		err := config.EnsureSshConfigEntryAdded(activeProfile.Id, workspace.Id, workspace.Projects[0].Name)
		if err != nil {
			return err
		}

		projectHostname := config.GetProjectHostname(activeProfile.Id, workspace.Id, workspace.Projects[0].Name)

		for {
			sshCommand := exec.Command("ssh", projectHostname, "daytona", "version")
			sshCommand.Stdin = nil
			sshCommand.Stdout = nil
			sshCommand.Stderr = &util.TraceLogWriter{}

			err = sshCommand.Run()
			if err == nil {
				return nil
			}

			time.Sleep(time.Second)
		}
	}

	for {
		dialConn, err := tsConn.Dial(context.Background(), "tcp", fmt.Sprintf("%s:%d", project.GetProjectHostname(workspace.Id, workspace.Projects[0].Name), ssh_config.SSH_PORT))
		if err == nil {
			return dialConn.Close()
		}

		time.Sleep(time.Second)
	}
}
