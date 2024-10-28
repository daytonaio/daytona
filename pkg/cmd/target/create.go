// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"
	"errors"
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
	target_util "github.com/daytonaio/daytona/pkg/cmd/target/util"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/target/workspace"
	"github.com/daytonaio/daytona/pkg/views"
	logs_view "github.com/daytonaio/daytona/pkg/views/logs"
	"github.com/daytonaio/daytona/pkg/views/target/create"
	"github.com/daytonaio/daytona/pkg/views/target/info"
	"github.com/daytonaio/daytona/pkg/views/target/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/docker/docker/pkg/stringid"
	log "github.com/sirupsen/logrus"
	"tailscale.com/tsnet"

	"github.com/spf13/cobra"

	"github.com/daytonaio/daytona/cmd/daytona/config"
)

var CreateCmd = &cobra.Command{
	Use:     "create [REPOSITORY_URL | WORKSPACE_CONFIG_NAME]...",
	Short:   "Create a target",
	GroupID: util.TARGET_GROUP,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		var workspaces []apiclient.CreateWorkspaceDTO
		var targetName string
		var existingTargetNames []string
		var existingWorkspaceConfigNames []string
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
			targetName = nameFlag
		}

		targetList, res, err := apiClient.TargetAPI.ListTargets(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}
		for _, targetInfo := range targetList {
			existingTargetNames = append(existingTargetNames, targetInfo.Name)
		}

		if promptUsingTUI {
			err = processPrompting(ctx, apiClient, &targetName, &workspaces, existingTargetNames)
			if err != nil {
				if common.IsCtrlCAbort(err) {
					return nil
				} else {
					return err
				}
			}
		} else {
			existingWorkspaceConfigNames, err = processCmdArguments(ctx, args, apiClient, &workspaces)
			if err != nil {
				return err
			}

			initialSuggestion := workspaces[0].Name

			if targetName == "" {
				targetName = target_util.GetSuggestedName(initialSuggestion, existingTargetNames)
			}
		}

		if targetName == "" || len(workspaces) == 0 {
			return errors.New("target name and repository urls are required")
		}

		workspaceNames := []string{}
		for i := range workspaces {
			if profileData != nil && profileData.EnvVars != nil {
				workspaces[i].EnvVars = util.MergeEnvVars(profileData.EnvVars, workspaces[i].EnvVars)
			} else {
				workspaces[i].EnvVars = util.MergeEnvVars(workspaces[i].EnvVars)
			}
			workspaceNames = append(workspaceNames, workspaces[i].Name)
		}

		for i, workspaceConfigName := range existingWorkspaceConfigNames {
			if workspaceConfigName == "" {
				continue
			}
			logs_view.DisplayLogEntry(logs.LogEntry{
				WorkspaceName: &workspaces[i].Name,
				Msg:           fmt.Sprintf("Using detected workspace config '%s'\n", workspaceConfigName),
			}, i)
		}

		targetConfigs, res, err := apiClient.TargetConfigAPI.ListTargetConfigs(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		targetConfig, err := target_util.GetTargetConfig(target_util.GetTargetConfigParams{
			Ctx:                  ctx,
			ApiClient:            apiClient,
			TargetConfigs:        targetConfigs,
			ActiveProfileName:    activeProfile.Name,
			TargetConfigNameFlag: targetConfigNameFlag,
			PromptUsingTUI:       promptUsingTUI,
		})
		if err != nil {
			if common.IsCtrlCAbort(err) {
				return nil
			}
			return err
		}

		logs_view.CalculateLongestPrefixLength(workspaceNames)

		logs_view.DisplayLogEntry(logs.LogEntry{
			Msg: "Request submitted\n",
		}, logs_view.STATIC_INDEX)

		activeProfile, err = c.GetActiveProfile()
		if err != nil {
			return err
		}

		var tsConn *tsnet.Server
		if targetConfig.Name != "local" || activeProfile.Id != "default" {
			tsConn, err = tailscale.GetConnection(&activeProfile)
			if err != nil {
				return err
			}
		}

		id := stringid.GenerateRandomID()
		id = stringid.TruncateID(id)

		logsContext, stopLogs := context.WithCancel(context.Background())
		go apiclient_util.ReadTargetLogs(logsContext, activeProfile, id, workspaceNames, true, true, nil)

		createdTarget, res, err := apiClient.TargetAPI.CreateTarget(ctx).Target(apiclient.CreateTargetDTO{
			Id:           id,
			Name:         targetName,
			TargetConfig: targetConfig.Name,
			Workspaces:   workspaces,
		}).Execute()
		if err != nil {
			stopLogs()
			return apiclient_util.HandleErrorResponse(res, err)
		}
		gpgKey, err := GetGitProviderGpgKey(apiClient, ctx, workspaces[0].GitProviderConfigId)
		if err != nil {
			log.Warn(err)
		}

		err = waitForDial(createdTarget, &activeProfile, tsConn, gpgKey)
		if err != nil {
			stopLogs()
			return err
		}

		stopLogs()

		// Make sure terminal cursor is reset
		fmt.Print("\033[?25h")

		targetInfo, res, err := apiClient.TargetAPI.GetTarget(ctx, targetName).Verbose(true).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
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
		info.Render(targetInfo, chosenIde.Name, false)

		if noIdeFlag {
			views.RenderCreationInfoMessage("Run 'daytona code' when you're ready to start developing")
			return nil
		}

		views.RenderCreationInfoMessage(fmt.Sprintf("Opening the target in %s ...", chosenIde.Name))

		workspaceName := targetInfo.Workspaces[0].Name
		providerMetadata, err := target_util.GetWorkspaceProviderMetadata(targetInfo, workspaceName)
		if err != nil {
			return err
		}

		return openIDE(chosenIdeId, activeProfile, createdTarget.Id, targetInfo.Workspaces[0].Name, providerMetadata, yesFlag, gpgKey)
	},
}

var nameFlag string
var targetConfigNameFlag string
var noIdeFlag bool
var blankFlag bool
var multiWorkspaceFlag bool

var workspaceConfigurationFlags = target_util.WorkspaceConfigurationFlags{
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

	CreateCmd.Flags().StringVar(&nameFlag, "name", "", "Specify the target name")
	CreateCmd.Flags().StringVarP(&ideFlag, "ide", "i", "", fmt.Sprintf("Specify the IDE (%s)", ideListStr))
	CreateCmd.Flags().StringVarP(&targetConfigNameFlag, "target", "t", "", "Specify the target (e.g. 'local')")
	CreateCmd.Flags().BoolVar(&blankFlag, "blank", false, "Create a blank workspace without using existing configurations")
	CreateCmd.Flags().BoolVarP(&noIdeFlag, "no-ide", "n", false, "Do not open the target in the IDE after target creation")
	CreateCmd.Flags().BoolVar(&multiWorkspaceFlag, "multi-workspace", false, "Target with multiple workspaces/repos")
	CreateCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Automatically confirm any prompts")
	CreateCmd.Flags().StringSliceVar(workspaceConfigurationFlags.Branches, "branch", []string{}, "Specify the Git branches to use in the workspaces")

	target_util.AddWorkspaceConfigurationFlags(CreateCmd, workspaceConfigurationFlags, true)
}

func processPrompting(ctx context.Context, apiClient *apiclient.APIClient, targetName *string, workspaces *[]apiclient.CreateWorkspaceDTO, targetNames []string) error {
	if target_util.CheckAnyWorkspaceConfigurationFlagSet(workspaceConfigurationFlags) || (workspaceConfigurationFlags.Branches != nil && len(*workspaceConfigurationFlags.Branches) > 0) {
		return errors.New("please provide the repository URL in order to set up custom workspace details through the CLI")
	}

	gitProviders, res, err := apiClient.GitProviderAPI.ListGitProviders(ctx).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}

	workspaceConfigs, res, err := apiClient.WorkspaceConfigAPI.ListWorkspaceConfigs(ctx).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}

	apiServerConfig, res, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}

	workspaceDefaults := &views_util.WorkspaceConfigDefaults{
		BuildChoice:          views_util.AUTOMATIC,
		Image:                &apiServerConfig.DefaultWorkspaceImage,
		ImageUser:            &apiServerConfig.DefaultWorkspaceUser,
		DevcontainerFilePath: create.DEVCONTAINER_FILEPATH,
	}

	*workspaces, err = target_util.GetWorkspacesCreationDataFromPrompt(target_util.WorkspacesDataPromptConfig{
		UserGitProviders: gitProviders,
		WorkspaceConfigs: workspaceConfigs,
		Manual:           *workspaceConfigurationFlags.Manual,
		MultiWorkspace:   multiWorkspaceFlag,
		BlankWorkspace:   blankFlag,
		ApiClient:        apiClient,
		Defaults:         workspaceDefaults,
	})

	if err != nil {
		return err
	}

	initialSuggestion := (*workspaces)[0].Name

	suggestedName := target_util.GetSuggestedName(initialSuggestion, targetNames)

	dedupWorkspaceNames(workspaces)

	submissionFormConfig := create.SubmissionFormConfig{
		ChosenName:    targetName,
		SuggestedName: suggestedName,
		ExistingNames: targetNames,
		WorkspaceList: workspaces,
		NameLabel:     "Target",
		Defaults:      workspaceDefaults,
	}

	pcImportFlag := false
	err = create.RunSubmissionForm(submissionFormConfig, &pcImportFlag)
	if err != nil {
		return err
	}

	return nil
}

func processCmdArguments(ctx context.Context, repoUrls []string, apiClient *apiclient.APIClient, workspaces *[]apiclient.CreateWorkspaceDTO) ([]string, error) {
	if len(repoUrls) == 0 {
		return nil, fmt.Errorf("no repository URLs provided")
	}

	if len(repoUrls) > 1 && target_util.CheckAnyWorkspaceConfigurationFlagSet(workspaceConfigurationFlags) {
		return nil, fmt.Errorf("can't set custom workspace configuration properties for multiple workspaces")
	}

	if *workspaceConfigurationFlags.Builder != "" && *workspaceConfigurationFlags.Builder != views_util.DEVCONTAINER && *workspaceConfigurationFlags.DevcontainerPath != "" {
		return nil, fmt.Errorf("can't set devcontainer file path if builder is not set to %s", views_util.DEVCONTAINER)
	}

	var workspaceConfig *apiclient.WorkspaceConfig

	existingWorkspaceConfigNames := []string{}

	for i, repoUrl := range repoUrls {
		var branch *string
		if len(*workspaceConfigurationFlags.Branches) > i {
			branch = &(*workspaceConfigurationFlags.Branches)[i]
		}

		validatedUrl, err := util.GetValidatedUrl(repoUrl)
		if err == nil {
			// The argument is a Git URL
			existingWorkspaceConfigName, err := processGitURL(ctx, validatedUrl, apiClient, workspaces, branch)
			if err != nil {
				return nil, err
			}
			if existingWorkspaceConfigName != nil {
				existingWorkspaceConfigNames = append(existingWorkspaceConfigNames, *existingWorkspaceConfigName)
			} else {
				existingWorkspaceConfigNames = append(existingWorkspaceConfigNames, "")
			}

			continue
		}

		// The argument is not a Git URL - try getting the workspace config
		workspaceConfig, _, err = apiClient.WorkspaceConfigAPI.GetWorkspaceConfig(ctx, repoUrl).Execute()
		if err != nil {
			return nil, fmt.Errorf("failed to parse the URL or fetch the workspace config for '%s'", repoUrl)
		}

		existingWorkspaceConfigName, err := target_util.AddWorkspaceFromConfig(workspaceConfig, apiClient, workspaces, branch)
		if err != nil {
			return nil, err
		}
		if existingWorkspaceConfigName != nil {
			existingWorkspaceConfigNames = append(existingWorkspaceConfigNames, *existingWorkspaceConfigName)
		} else {
			existingWorkspaceConfigNames = append(existingWorkspaceConfigNames, "")
		}
	}

	dedupWorkspaceNames(workspaces)

	return existingWorkspaceConfigNames, nil
}

func processGitURL(ctx context.Context, repoUrl string, apiClient *apiclient.APIClient, workspaces *[]apiclient.CreateWorkspaceDTO, branch *string) (*string, error) {
	encodedURLParam := url.QueryEscape(repoUrl)

	if !blankFlag {
		workspaceConfig, res, err := apiClient.WorkspaceConfigAPI.GetDefaultWorkspaceConfig(ctx, encodedURLParam).Execute()
		if err == nil {
			workspaceConfig.GitProviderConfigId = workspaceConfigurationFlags.GitProviderConfig
			return target_util.AddWorkspaceFromConfig(workspaceConfig, apiClient, workspaces, branch)
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

	workspaceName, err := target_util.GetSanitizedWorkspaceName(repo.Name)
	if err != nil {
		return nil, err
	}

	workspaceConfigurationFlags.GitProviderConfig, err = target_util.GetGitProviderConfigIdFromFlag(ctx, apiClient, workspaceConfigurationFlags.GitProviderConfig)
	if err != nil {
		return nil, err
	}

	if workspaceConfigurationFlags.GitProviderConfig == nil || *workspaceConfigurationFlags.GitProviderConfig == "" {
		gitProviderConfigs, res, err := apiClient.GitProviderAPI.ListGitProvidersForUrl(context.Background(), url.QueryEscape(repoUrl)).Execute()
		if err != nil {
			return nil, apiclient_util.HandleErrorResponse(res, err)
		}

		if len(gitProviderConfigs) == 1 {
			workspaceConfigurationFlags.GitProviderConfig = &gitProviderConfigs[0].Id
		} else if len(gitProviderConfigs) > 1 {
			gp := selection.GetGitProviderConfigFromPrompt(selection.GetGitProviderConfigParams{
				GitProviderConfigs: gitProviderConfigs,
				ActionVerb:         "Use",
			})
			if gp == nil {
				return nil, common.ErrCtrlCAbort
			}
			workspaceConfigurationFlags.GitProviderConfig = &gp.Id
		}
	}

	workspace, err := target_util.GetCreateWorkspaceDtoFromFlags(workspaceConfigurationFlags)
	if err != nil {
		return nil, err
	}

	workspace.Name = workspaceName
	workspace.Source = apiclient.CreateWorkspaceSourceDTO{
		Repository: *repo,
	}

	*workspaces = append(*workspaces, *workspace)

	return nil, nil
}

func waitForDial(target *apiclient.Target, activeProfile *config.Profile, tsConn *tsnet.Server, gpgKey string) error {
	if target.TargetConfig == "local" && (activeProfile != nil && activeProfile.Id == "default") {
		err := config.EnsureSshConfigEntryAdded(activeProfile.Id, target.Id, target.Workspaces[0].Name, gpgKey)
		if err != nil {
			return err
		}

		workspaceHostname := config.GetWorkspaceHostname(activeProfile.Id, target.Id, target.Workspaces[0].Name)

		for {
			sshCommand := exec.Command("ssh", workspaceHostname, "daytona", "version")
			sshCommand.Stdin = nil
			sshCommand.Stdout = nil
			sshCommand.Stderr = &util.TraceLogWriter{}

			err = sshCommand.Run()
			if err == nil {
				return nil
			}

			time.Sleep(100 * time.Millisecond)
		}
	}

	connectChan := make(chan error)
	spinner := time.After(15 * time.Second)
	timeout := time.After(2 * time.Minute)

	go func() {
		for {
			dialConn, err := tsConn.Dial(context.Background(), "tcp", fmt.Sprintf("%s:%d", workspace.GetWorkspaceHostname(target.Id, target.Workspaces[0].Name), ssh_config.SSH_PORT))
			if err == nil {
				connectChan <- dialConn.Close()
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	select {
	case err := <-connectChan:
		return err
	case <-spinner:
		err := views_util.WithInlineSpinner("Connection to tailscale is taking longer than usual", func() error {
			select {
			case err := <-connectChan:
				return err
			case <-timeout:
				return errors.New("secure connection to the Daytona Server could not be established. Please check your internet connection or Tailscale availability")
			}
		})
		return err
	}
}

func dedupWorkspaceNames(workspaces *[]apiclient.CreateWorkspaceDTO) {
	workspaceNames := map[string]int{}

	for i, workspace := range *workspaces {
		if _, ok := workspaceNames[workspace.Name]; ok {
			(*workspaces)[i].Name = fmt.Sprintf("%s-%d", workspace.Name, workspaceNames[workspace.Name])
			workspaceNames[workspace.Name]++
		} else {
			workspaceNames[workspace.Name] = 2
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
