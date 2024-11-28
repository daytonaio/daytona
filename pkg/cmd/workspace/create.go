// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/daytonaio/daytona/internal/cmd/tailscale"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	workspace_util "github.com/daytonaio/daytona/pkg/cmd/workspace/util"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/views"
	logs_view "github.com/daytonaio/daytona/pkg/views/logs"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/create"
	"github.com/daytonaio/daytona/pkg/views/workspace/info"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	"github.com/docker/docker/pkg/stringid"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"

	"github.com/daytonaio/daytona/cmd/daytona/config"
)

var (
  nameFlag          string
  targetNameFlag    string
  noIdeFlag         bool
  blankFlag         bool
  multiProjectFlag  bool
  createIdeFlag     string // Ensure this is declared only once
  createYesFlag     bool   // Ensure this is declared only once
)

var CreateCmd = &cobra.Command{
  Use:    "create [REPOSITORY_URL | PROJECT_CONFIG_NAME]...",
  Short:  "Create a workspace",
  GroupID: util.WORKSPACE_GROUP,
  RunE: func(cmd *cobra.Command, args []string) error {
  	ctx := context.Background()
  	var projects []apiclient.CreateProjectDTO
  	var workspaceName string
  	var existingWorkspaceNames []string
  	var existingProjectConfigNames []string
  	promptUsingTUI := len(args) == 0

  	apiClient, err := apiclient_util.GetApiClient(nil)
  	if err != nil {
  		return err
  	}

  	c, err := config.GetConfig()
  	if err != nil {
  		return err
  	}

  	activeProfile, err := c.GetActiveProfile()
  	if err != nil {
  		return err
  	}

  	profileData, res, err := apiClient.ProfileAPI.GetProfileData(ctx).Execute()
  	if err != nil {
  		return apiclient_util.HandleErrorResponse(res, err)
  	}

  	if nameFlag != "" {
  		workspaceName = nameFlag
  	}

  	workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
  	if err != nil {
  		return apiclient_util.HandleErrorResponse(res, err)
  	}
  	for _, workspaceInfo := range workspaceList {
  		existingWorkspaceNames = append(existingWorkspaceNames, workspaceInfo.Name)
  	}

  	if promptUsingTUI {
  		err = processPrompting(ctx, apiClient, &workspaceName, &projects, existingWorkspaceNames)
  		if err != nil {
  			if common.IsCtrlCAbort(err) {
  				return nil
  			}
  			return err
  		}
  	} else {
  		existingProjectConfigNames, err = processCmdArguments(ctx, args, apiClient, &projects)
  		if err != nil {
  			return err
  		}

  		initialSuggestion := projects[0].Name

  		if workspaceName == "" {
  			workspaceName = workspace_util.GetSuggestedName(initialSuggestion, existingWorkspaceNames)
  		}
  	}

  	if workspaceName == "" || len(projects) == 0 {
  		return errors.New("workspace name and repository URLs are required")
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

  	for i, projectConfigName := range existingProjectConfigNames {
  		if projectConfigName == "" {
  			continue
  		}
  		logs_view.DisplayLogEntry(logs.LogEntry{
  			ProjectName: &projects[i].Name,
  			Msg:         fmt.Sprintf("Using detected project config '%s'\n", projectConfigName),
  		}, i)
  	}

  	targetList, res, err := apiClient.TargetAPI.ListTargets(ctx).Execute()
  	if err != nil {
  		return apiclient_util.HandleErrorResponse(res, err)
  	}

  	target, err := workspace_util.GetTarget(workspace_util.GetTargetConfig{
  		Ctx:               ctx,
  		ApiClient:         apiClient,
  		TargetList:        targetList,
  		ActiveProfileName: activeProfile.Name,
  		TargetNameFlag:    targetNameFlag,
  		PromptUsingTUI:    promptUsingTUI,
  	})
  	if err != nil {
  		if common.IsCtrlCAbort(err) {
  			return nil
  		}
  		return err
  	}

  	logs_view.CalculateLongestPrefixLength(projectNames)

  	logs_view.DisplayLogEntry(logs.LogEntry{
  		Msg: "Request submitted\n",
  	}, logs_view.STATIC_INDEX)

  	activeProfile, err = c.GetActiveProfile()
  	if err != nil {
  		return err
  	}

  	// Remove tsConn if not used
  	// var tsConn *tsnet.Server
  	if target.Name != "local" || activeProfile.Id != "default" {
  		_, err = tailscale.GetConnection(&activeProfile)
  		if err != nil {
  			return err
  		}
  	}

  	id := stringid.GenerateRandomID()
  	id = stringid.TruncateID(id)

  	logsContext, stopLogs := context.WithCancel(context.Background())
  	go func() {
  		apiclient_util.ReadWorkspaceLogs(logsContext, activeProfile, id, projectNames, true, true, nil)
  	}()

  	createdWorkspace, res, err := apiClient.WorkspaceAPI.CreateWorkspace(ctx).Workspace(apiclient.CreateWorkspaceDTO{
  		Id:       id,
  		Name:     workspaceName,
  		Target:   target.Name,
  		Projects: projects,
  	}).Execute()
  	if err != nil {
  		stopLogs()
  		return apiclient_util.HandleErrorResponse(res, err)
  	}
  	gpgKey, err := GetGitProviderGpgKey(apiClient, ctx, projects[0].GitProviderConfigId)
  	if err != nil {
  		log.Warn(err)
  	}

  	// Replace the existing log reading logic with a retry mechanism
  	err = readLogsWithRetry(ctx, activeProfile, createdWorkspace.Id, projectNames)
  	if err != nil {
  		log.Warnf("Error reading logs: %v", err)
  		log.Info("Workspace creation process is continuing on the server.")
  		log.Info("Use 'daytona logs' command to check the status later.")
  	}

  	stopLogs()

  	// Make sure terminal cursor is reset
  	fmt.Print("\033[?25h")

  	wsInfo, res, err := apiClient.WorkspaceAPI.GetWorkspace(ctx, workspaceName).Verbose(true).Execute()
  	if err != nil {
  		return apiclient_util.HandleErrorResponse(res, err)
  	}

  	chosenIdeId := c.DefaultIdeId
  	if createIdeFlag != "" {
  		chosenIdeId = createIdeFlag
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

  	if noIdeFlag {
  		views.RenderCreationInfoMessage("Run 'daytona code' when you're ready to start developing")
  		return nil
  	}

  	views.RenderCreationInfoMessage(fmt.Sprintf("Opening the workspace in %s ...", chosenIde.Name))

  	projectName := wsInfo.Projects[0].Name
  	providerMetadata, err := workspace_util.GetProjectProviderMetadata(wsInfo, projectName)
  	if err != nil {
  		return err
  	}

  	return openIDE(chosenIdeId, activeProfile, createdWorkspace.Id, wsInfo.Projects[0].Name, providerMetadata, createYesFlag, gpgKey)
  },
}

var projectConfigurationFlags = workspace_util.ProjectConfigurationFlags{
  Builder:           new(views_util.BuildChoice),
  CustomImage:       new(string),
  CustomImageUser:   new(string),
  Branches:          new([]string),
  DevcontainerPath:  new(string),
  EnvVars:           new([]string),
  Manual:            new(bool),
  GitProviderConfig: new(string),
}

func init() {
  ideList := config.GetIdeList()
  ids := make([]string, len(ideList))
  for i, ide := range ideList {
  	ids[i] = ide.Id
  }
  ideListStr := strings.Join(ids, ", ")

  CreateCmd.Flags().StringVar(&nameFlag, "name", "", "Specify the workspace name")
  CreateCmd.Flags().StringVarP(&createIdeFlag, "ide", "i", "", fmt.Sprintf("Specify the IDE (%s)", ideListStr))
  CreateCmd.Flags().StringVarP(&targetNameFlag, "target", "t", "", "Specify the target (e.g. 'local')")
  CreateCmd.Flags().BoolVar(&blankFlag, "blank", false, "Create a blank project without using existing configurations")
  CreateCmd.Flags().BoolVarP(&noIdeFlag, "no-ide", "n", false, "Do not open the workspace in the IDE after workspace creation")
  CreateCmd.Flags().BoolVar(&multiProjectFlag, "multi-project", false, "Workspace with multiple projects/repos")
  CreateCmd.Flags().BoolVarP(&createYesFlag, "yes", "y", false, "Automatically confirm any prompts")
  CreateCmd.Flags().StringSliceVar(projectConfigurationFlags.Branches, "branch", []string{}, "Specify the Git branches to use in the projects")

  workspace_util.AddProjectConfigurationFlags(CreateCmd, projectConfigurationFlags, true)
}

func processPrompting(ctx context.Context, apiClient *apiclient.APIClient, workspaceName *string, projects *[]apiclient.CreateProjectDTO, workspaceNames []string) error {
  if workspace_util.CheckAnyProjectConfigurationFlagSet(projectConfigurationFlags) || (projectConfigurationFlags.Branches != nil && len(*projectConfigurationFlags.Branches) > 0) {
  	return errors.New("please provide the repository URL in order to set up custom project details through the CLI")
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

  dedupProjectNames(projects)

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

func processCmdArguments(ctx context.Context, repoUrls []string, apiClient *apiclient.APIClient, projects *[]apiclient.CreateProjectDTO) ([]string, error) {
  if len(repoUrls) == 0 {
  	return nil, fmt.Errorf("no repository URLs provided")
  }

  if len(repoUrls) > 1 && workspace_util.CheckAnyProjectConfigurationFlagSet(projectConfigurationFlags) {
  	return nil, fmt.Errorf("can't set custom project configuration properties for multiple projects")
  }

  if *projectConfigurationFlags.Builder != "" && *projectConfigurationFlags.Builder != views_util.DEVCONTAINER && *projectConfigurationFlags.DevcontainerPath != "" {
  	return nil, fmt.Errorf("can't set devcontainer file path if builder is not set to %s", views_util.DEVCONTAINER)
  }

  var projectConfig *apiclient.ProjectConfig

  existingProjectConfigNames := []string{}

  for i, repoUrl := range repoUrls {
  	var branch *string
  	if len(*projectConfigurationFlags.Branches) > i {
  		branch = &(*projectConfigurationFlags.Branches)[i]
  	}

  	validatedUrl, err := util.GetValidatedUrl(repoUrl)
  	if err == nil {
  		// The argument is a Git URL
  		existingProjectConfigName, err := processGitURL(ctx, validatedUrl, apiClient, projects, branch)
  		if err != nil {
  			return nil, err
  		}
  		if existingProjectConfigName != nil {
  			existingProjectConfigNames = append(existingProjectConfigNames, *existingProjectConfigName)
  		} else {
  			existingProjectConfigNames = append(existingProjectConfigNames, "")
  		}

  		continue
  	}

  	// The argument is not a Git URL - try getting the project config
  	projectConfig, _, err = apiClient.ProjectConfigAPI.GetProjectConfig(ctx, repoUrl).Execute()
  	if err != nil {
  		return nil, fmt.Errorf("failed to parse the URL or fetch the project config for '%s'", repoUrl)
  	}

  	existingProjectConfigName, err := workspace_util.AddProjectFromConfig(projectConfig, apiClient, projects, branch)
  	if err != nil {
  		return nil, err
  	}
  	if existingProjectConfigName != nil {
  		existingProjectConfigNames = append(existingProjectConfigNames, *existingProjectConfigName)
  	} else {
  		existingProjectConfigNames = append(existingProjectConfigNames, "")
  	}
  }

  dedupProjectNames(projects)

  return existingProjectConfigNames, nil
}

func processGitURL(ctx context.Context, repoUrl string, apiClient *apiclient.APIClient, projects *[]apiclient.CreateProjectDTO, branch *string) (*string, error) {
  encodedURLParam := url.QueryEscape(repoUrl)

  if !blankFlag {
  	projectConfig, res, err := apiClient.ProjectConfigAPI.GetDefaultProjectConfig(ctx, encodedURLParam).Execute()
  	if err == nil {
  		projectConfig.GitProviderConfigId = projectConfigurationFlags.GitProviderConfig
  		return workspace_util.AddProjectFromConfig(projectConfig, apiClient, projects, branch)
  	}

  	if res.StatusCode != http.StatusNotFound {
  		return nil, apiclient_util.HandleErrorResponse(res, err)
  	}
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

  projectConfigurationFlags.GitProviderConfig, err = workspace_util.GetGitProviderConfigIdFromFlag(ctx, apiClient, projectConfigurationFlags.GitProviderConfig)
  if err != nil {
  	return nil, err
  }

  if projectConfigurationFlags.GitProviderConfig == nil || *projectConfigurationFlags.GitProviderConfig == "" {
  	gitProviderConfigs, res, err := apiClient.GitProviderAPI.ListGitProvidersForUrl(context.Background(), url.QueryEscape(repoUrl)).Execute()
  	if err != nil {
  		return nil, apiclient_util.HandleErrorResponse(res, err)
  	}

  	if len(gitProviderConfigs) == 1 {
  		projectConfigurationFlags.GitProviderConfig = &gitProviderConfigs[0].Id
  	} else if len(gitProviderConfigs) > 1 {
  		gp := selection.GetGitProviderConfigFromPrompt(selection.GetGitProviderConfigParams{
  			GitProviderConfigs: gitProviderConfigs,
  			ActionVerb:         "Use",
  		})
  		if gp == nil {
  			return nil, common.ErrCtrlCAbort
  		}
  		projectConfigurationFlags.GitProviderConfig = &gp.Id
  	}
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

func dedupProjectNames(projects *[]apiclient.CreateProjectDTO) {
  projectNames := map[string]int{}

  for i, project := range *projects {
  	if _, ok := projectNames[project.Name]; ok {
  		(*projects)[i].Name = fmt.Sprintf("%s-%d", project.Name, projectNames[project.Name])
  		projectNames[project.Name]++
  	} else {
  		projectNames[project.Name] = 2
  	}
  }
}

func GetGitProviderGpgKey(apiClient *apiclient.APIClient, ctx context.Context, providerConfigId *string) (string, error) {
  if providerConfigId == nil || *providerConfigId == "" {
  	return "", nil
  }

  var providerConfig *gitprovider.GitProviderConfig
  var gpgKey string

  gitProvider, res, err := apiClient.GitProviderAPI.GetGitProvider(ctx, *providerConfigId).Execute()
  if err != nil {
  	return "", apiclient_util.HandleErrorResponse(res, err)
  }

  // Extract GPG key if present
  if gitProvider != nil {
  	providerConfig = &gitprovider.GitProviderConfig{
  		SigningMethod: (*gitprovider.SigningMethod)(gitProvider.SigningMethod),
  		SigningKey:    gitProvider.SigningKey,
  	}

  	if providerConfig.SigningMethod != nil && providerConfig.SigningKey != nil {
  		if *providerConfig.SigningMethod == gitprovider.SigningMethodGPG {
  			gpgKey = *providerConfig.SigningKey
  		}
  	}
  }

  return gpgKey, nil
}

func readLogsWithRetry(ctx context.Context, activeProfile config.Profile, workspaceID string, projectNames []string) error {
  maxRetries := 5
  retryDelay := time.Second

  for retry := 0; retry < maxRetries; retry++ {
  	logsContext, stopLogs := context.WithCancel(ctx)
  	defer stopLogs()

  	errChan := make(chan error, 1)
  	go func() {
  		apiclient_util.ReadWorkspaceLogs(logsContext, activeProfile, workspaceID, projectNames, true, true, nil)
  		errChan <- nil
  	}()

  	select {
  	case err := <-errChan:
  		if err == nil {
  			return nil
  		}
  		if retry < maxRetries-1 {
  			log.Warnf("Log read error: %v. Retrying in %v...", err, retryDelay)
  			time.Sleep(retryDelay)
  			retryDelay *= 2 // Exponential backoff
  		}
  	case <-ctx.Done():
  		return ctx.Err()
  	}
  }

  return fmt.Errorf("failed to read logs after %d attempts", maxRetries)
}