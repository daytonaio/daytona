// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
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
		var projects []apiclient.CreateProjectConfigDTO
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
			existingWorkspaceNames = append(existingWorkspaceNames, *workspaceInfo.Name)
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

			initialSuggestion := *projects[0].Name

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
			projects[i].EnvVars = workspace_util.GetEnvVariables(&projects[i], profileData)
			projectNames = append(projectNames, *projects[i].Name)
		}

		logs_view.CalculateLongestPrefixLength(projectNames)

		logs_view.DisplayLogEntry(logs.LogEntry{
			Msg: "Request submitted\n",
		}, logs_view.WORKSPACE_INDEX)

		if existingProjectConfigName != nil {
			logs_view.DisplayLogEntry(logs.LogEntry{
				ProjectName: *existingProjectConfigName,
				Msg:         fmt.Sprintf("Using detected project config '%s'\n", *existingProjectConfigName),
			}, logs_view.FIRST_PROJECT_INDEX)
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

		go apiclient_util.ReadWorkspaceLogs(activeProfile, id, projectNames, &stopLogs)

		createdWorkspace, res, err := apiClient.WorkspaceAPI.CreateWorkspace(ctx).Workspace(apiclient.CreateWorkspaceDTO{
			Id:       &id,
			Name:     &workspaceName,
			Target:   target.Name,
			Projects: projects,
		}).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		err = waitForDial(tsConn, *createdWorkspace.Id, *createdWorkspace.Projects[0].Name)
		if err != nil {
			log.Fatal(err)
		}

		stopLogs = true

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

		providerMetadata := ""
		for _, project := range wsInfo.Info.Projects {
			if *project.Name == *wsInfo.Projects[0].Name {
				if project.ProviderMetadata == nil {
					log.Fatal(errors.New("project provider metadata is missing"))
				}
				providerMetadata = *project.ProviderMetadata
				break
			}
		}

		err = openIDE(chosenIdeId, activeProfile, *createdWorkspace.Id, *wsInfo.Projects[0].Name, providerMetadata)
		if err != nil {
			log.Fatal(err)
		}
	},
}

var providerFlag string
var nameFlag string
var targetNameFlag string
var customImageFlag string
var customImageUserFlag string
var devcontainerPathFlag string
var branchFlag string

var builderFlag create.BuildChoice

var manualFlag bool
var multiProjectFlag bool
var codeFlag bool
var blankFlag bool

func init() {
	CreateCmd.Flags().StringVar(&nameFlag, "name", "", "Specify the workspace name")
	CreateCmd.Flags().StringVar(&providerFlag, "provider", "", "Specify the provider (e.g. 'docker-provider')")
	CreateCmd.Flags().StringVarP(&ideFlag, "ide", "i", "", "Specify the IDE ('vscode' or 'browser')")
	CreateCmd.Flags().StringVarP(&targetNameFlag, "target", "t", "", "Specify the target (e.g. 'local')")
	CreateCmd.Flags().StringVar(&customImageFlag, "custom-image", "", "Create the project with the custom image passed as the flag value; Requires setting --custom-image-user flag as well")
	CreateCmd.Flags().StringVar(&customImageUserFlag, "custom-image-user", "", "Create the project with the custom image user passed as the flag value; Requires setting --custom-image flag as well")
	CreateCmd.Flags().StringVar(&branchFlag, "branch", "", "Specify the Git branch to use in the project")
	CreateCmd.Flags().StringVar(&devcontainerPathFlag, "devcontainer-path", "", "Automatically assign the devcontainer builder with the path passed as the flag value")

	CreateCmd.Flags().Var(&builderFlag, "builder", fmt.Sprintf("Specify the builder (currently %s/%s/%s)", create.AUTOMATIC, create.DEVCONTAINER, create.NONE))

	CreateCmd.Flags().BoolVar(&manualFlag, "manual", false, "Manually enter the Git repositories")
	CreateCmd.Flags().BoolVar(&multiProjectFlag, "multi-project", false, "Workspace with multiple projects/repos")
	CreateCmd.Flags().BoolVar(&blankFlag, "blank", false, "Create a blank project without using existing configurations")
	CreateCmd.Flags().BoolVarP(&codeFlag, "code", "c", false, "Open the workspace in the IDE after workspace creation")

	CreateCmd.MarkFlagsMutuallyExclusive("multi-project", "custom-image")
	CreateCmd.MarkFlagsMutuallyExclusive("multi-project", "custom-image-user")
	CreateCmd.MarkFlagsMutuallyExclusive("multi-project", "devcontainer-path")
	CreateCmd.MarkFlagsMutuallyExclusive("multi-project", "builder")
	CreateCmd.MarkFlagsMutuallyExclusive("builder", "custom-image")
	CreateCmd.MarkFlagsMutuallyExclusive("builder", "custom-image-user")
	CreateCmd.MarkFlagsMutuallyExclusive("devcontainer-path", "custom-image")
	CreateCmd.MarkFlagsMutuallyExclusive("devcontainer-path", "custom-image-user")

	CreateCmd.MarkFlagsRequiredTogether("custom-image", "custom-image-user")
}

func getTarget(activeProfileName string) (*apiclient.ProviderTarget, error) {
	targets, err := apiclient_util.GetTargetList()
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

func processPrompting(apiClient *apiclient.APIClient, workspaceName *string, projects *[]apiclient.CreateProjectConfigDTO, workspaceNames []string, ctx context.Context) error {
	if builderFlag != "" || customImageFlag != "" || customImageUserFlag != "" || devcontainerPathFlag != "" {
		return fmt.Errorf("please provide the repository URL in order to set up custom project details through CLI")
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

	projectDefaults := &create.ProjectConfigDefaults{
		BuildChoice:          create.AUTOMATIC,
		Image:                apiServerConfig.DefaultProjectImage,
		ImageUser:            apiServerConfig.DefaultProjectUser,
		DevcontainerFilePath: create.DEVCONTAINER_FILEPATH,
	}

	*projects, err = workspace_util.GetProjectsCreationDataFromPrompt(workspace_util.ProjectsDataPromptConfig{
		UserGitProviders: gitProviders,
		ProjectConfigs:   projectConfigs,
		Manual:           manualFlag,
		MultiProject:     multiProjectFlag,
		BlankProject:     blankFlag,
		ApiClient:        apiClient,
		Defaults:         projectDefaults,
	},
	)
	if err != nil {
		return err
	}

	initialSuggestion := *(*projects)[0].Name

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

func processCmdArgument(argument string, apiClient *apiclient.APIClient, projects *[]apiclient.CreateProjectConfigDTO, ctx context.Context) (*string, error) {
	if builderFlag != "" && builderFlag != create.DEVCONTAINER && devcontainerPathFlag != "" {
		return nil, fmt.Errorf("can't set devcontainer file path if builder is not set to %s", create.DEVCONTAINER)
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

	return workspace_util.AddProjectFromConfig(projectConfig, apiClient, projects, branchFlag)
}

func processGitURL(repoUrl string, apiClient *apiclient.APIClient, projects *[]apiclient.CreateProjectConfigDTO, ctx context.Context) (*string, error) {
	encodedURLParam := url.QueryEscape(repoUrl)

	if !blankFlag {
		projectConfig, res, err := apiClient.ProjectConfigAPI.GetDefaultProjectConfig(ctx, encodedURLParam).Execute()
		if err == nil {
			return workspace_util.AddProjectFromConfig(projectConfig, apiClient, projects, branchFlag)
		}

		if res.StatusCode != http.StatusNotFound {
			return nil, apiclient_util.HandleErrorResponse(res, err)
		}
	}

	repoResponse, res, err := apiClient.GitProviderAPI.GetGitContext(ctx, encodedURLParam).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	if branchFlag != "" {
		repoResponse.Branch = &branchFlag
	}

	projectName, err := workspace_util.GetSanitizedProjectName(*repoResponse.Name)
	if err != nil {
		return nil, err
	}

	project := &apiclient.CreateProjectConfigDTO{
		Name: &projectName,
		Source: &apiclient.CreateProjectConfigSourceDTO{
			Repository: repoResponse,
		},
		BuildConfig: &apiclient.ProjectBuildConfig{},
	}

	if builderFlag == create.DEVCONTAINER || devcontainerPathFlag != "" {
		devcontainerFilePath := create.DEVCONTAINER_FILEPATH
		if devcontainerPathFlag != "" {
			devcontainerFilePath = devcontainerPathFlag
		}
		project.BuildConfig.Devcontainer = &apiclient.DevcontainerConfig{
			FilePath: &devcontainerFilePath,
		}

	}

	if builderFlag == create.NONE || customImageFlag != "" || customImageUserFlag != "" {
		project.BuildConfig = nil
		if customImageFlag != "" || customImageUserFlag != "" {
			project.Image = &customImageFlag
			project.User = &customImageUserFlag
		}
	}

	*projects = append(*projects, *project)

	return nil, nil
}

func waitForDial(tsConn *tsnet.Server, workspaceId string, projectName string) error {
	for {
		dialConn, err := tsConn.Dial(context.Background(), "tcp", fmt.Sprintf("%s:%d", project.GetProjectHostname(workspaceId, projectName), ssh_config.SSH_PORT))
		if err == nil {
			defer dialConn.Close()
			break
		}

		time.Sleep(time.Second)
	}
	return nil
}
