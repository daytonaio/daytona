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
	"time"

	"github.com/daytonaio/daytona/internal/cmd/tailscale"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	ssh_config "github.com/daytonaio/daytona/pkg/agent/ssh/config"
	"github.com/daytonaio/daytona/pkg/apiclient"
	workspace_util "github.com/daytonaio/daytona/pkg/cmd/workspace/util"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/views"
	logs_view "github.com/daytonaio/daytona/pkg/views/logs"
	"github.com/daytonaio/daytona/pkg/views/target"
	"github.com/daytonaio/daytona/pkg/views/workspace/create"
	"github.com/daytonaio/daytona/pkg/views/workspace/info"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/docker/docker/pkg/stringid"
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
		var projects []apiclient.CreateWorkspaceRequestProject
		var workspaceName string
		var existingWorkspaceNames []string

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

			if builderFlag == create.DEVCONTAINER || devcontainerPathFlag != "" {
				projects[i].Build = &apiclient.ProjectBuild{
					Devcontainer: &apiclient.ProjectBuildDevcontainer{},
				}
				if devcontainerPathFlag != "" {
					projects[i].Build.Devcontainer.DevContainerFilePath = &devcontainerPathFlag
				}
			}

			if builderFlag == create.AUTOMATIC {
				projects[i].Build = &apiclient.ProjectBuild{}
			}

			if builderFlag == create.NONE {
				projects[i].Build = nil
			}

			if customImageFlag != "" || customImageUserFlag != "" {
				projects[i].Build = nil
				projects[i].Image = &customImageFlag
				projects[i].User = &customImageUserFlag
			}
		}

		projectNames := []string{}
		for _, project := range projects {
			projectNames = append(projectNames, project.Name)
		}

		logs_view.CalculateLongestPrefixLength(projectNames)

		requestSubmittedLog := logs.LogEntry{
			Msg: "Request submitted\n",
		}

		logs_view.DisplayLogEntry(requestSubmittedLog, logs_view.WORKSPACE_INDEX)

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

		createdWorkspace, res, err := apiClient.WorkspaceAPI.CreateWorkspace(ctx).Workspace(apiclient.CreateWorkspaceRequest{
			Id:       &id,
			Name:     &workspaceName,
			Target:   target.Name,
			Projects: projects,
		}).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
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
var customImageFlag string
var customImageUserFlag string
var devcontainerPathFlag string

var builderFlag create.BuildChoice

var manualFlag bool
var multiProjectFlag bool
var codeFlag bool

func init() {
	CreateCmd.Flags().StringVar(&nameFlag, "name", "", "Specify the workspace name")
	CreateCmd.Flags().StringVar(&providerFlag, "provider", "", "Specify the provider (e.g. 'docker-provider')")
	CreateCmd.Flags().StringVarP(&ideFlag, "ide", "i", "", "Specify the IDE ('vscode' or 'browser')")
	CreateCmd.Flags().StringVarP(&targetNameFlag, "target", "t", "", "Specify the target (e.g. 'local')")
	CreateCmd.Flags().StringVar(&customImageFlag, "custom-image", "", "Create the project with the custom image passed as the flag value; Requires setting --custom-image-user flag as well")
	CreateCmd.Flags().StringVar(&customImageUserFlag, "custom-image-user", "", "Create the project with the custom image user passed as the flag value; Requires setting --custom-image flag as well")
	CreateCmd.Flags().StringVar(&devcontainerPathFlag, "devcontainer-path", "", "Automatically assign the devcontainer builder with the path passed as the flag value")

	CreateCmd.Flags().Var(&builderFlag, "builder", fmt.Sprintf("Specify the builder (currently %s/%s/%s)", create.AUTOMATIC, create.DEVCONTAINER, create.NONE))

	CreateCmd.Flags().BoolVar(&manualFlag, "manual", false, "Manually enter the git repositories")
	CreateCmd.Flags().BoolVar(&multiProjectFlag, "multi-project", false, "Workspace with multiple projects/repos")
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

func processPrompting(apiClient *apiclient.APIClient, workspaceName *string, projects *[]apiclient.CreateWorkspaceRequestProject, workspaceNames []string, ctx context.Context) error {
	gitProviders, res, err := apiClient.GitProviderAPI.ListGitProviders(ctx).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}

	apiServerConfig, res, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
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

func processCmdArguments(args []string, apiClient *apiclient.APIClient, projects *[]apiclient.CreateWorkspaceRequestProject, ctx context.Context) error {
	if builderFlag != "" && builderFlag != create.DEVCONTAINER && devcontainerPathFlag != "" {
		return fmt.Errorf("Can't set devcontainer file path if builder is not set to %s.", create.DEVCONTAINER)
	}

	repoUrl := args[0]

	repoUrl, err := util.GetValidatedUrl(repoUrl)
	if err != nil {
		return err
	}

	encodedURLParam := url.QueryEscape(repoUrl)
	repoResponse, res, err := apiClient.GitProviderAPI.GetGitContext(ctx, encodedURLParam).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}

	projectName, err := workspace_util.GetSanitizedProjectName(*repoResponse.Name)
	if err != nil {
		return err
	}

	project := &apiclient.CreateWorkspaceRequestProject{
		Name: projectName,
		Source: &apiclient.CreateWorkspaceRequestProjectSource{
			Repository: repoResponse,
		},
		Build: &apiclient.ProjectBuild{},
	}

	*projects = append(*projects, *project)

	return nil
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

func getEnvVariables(project *apiclient.CreateWorkspaceRequestProject, profileData *apiclient.ProfileData) *map[string]string {
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
