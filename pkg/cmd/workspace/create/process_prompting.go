// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"context"
	"errors"

	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	workspace_common "github.com/daytonaio/daytona/pkg/cmd/workspace/common"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/create"
)

type ProcessPromptingParams struct {
	ApiClient                   *apiclient.APIClient
	CreateWorkspaceDtos         *[]apiclient.CreateWorkspaceDTO
	ExistingWorkspaces          *[]apiclient.WorkspaceDTO
	WorkspaceConfigurationFlags workspace_common.WorkspaceConfigurationFlags
	MultiWorkspaceFlag          bool
	BlankFlag                   bool
	TargetName                  string
}

func ProcessPrompting(ctx context.Context, config ProcessPromptingParams) error {
	if workspace_common.CheckAnyWorkspaceConfigurationFlagSet(config.WorkspaceConfigurationFlags) || (config.WorkspaceConfigurationFlags.Branches != nil && len(*config.WorkspaceConfigurationFlags.Branches) > 0) {
		return errors.New("please provide the repository URL in order to set up custom workspace details through the CLI")
	}

	gitProviders, res, err := config.ApiClient.GitProviderAPI.ListGitProviders(ctx).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}

	workspaceConfigs, res, err := config.ApiClient.WorkspaceConfigAPI.ListWorkspaceConfigs(ctx).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}

	apiServerConfig, res, err := config.ApiClient.ServerAPI.GetConfig(ctx).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}

	workspaceDefaults := &views_util.WorkspaceConfigDefaults{
		BuildChoice:          views_util.AUTOMATIC,
		Image:                &apiServerConfig.DefaultWorkspaceImage,
		ImageUser:            &apiServerConfig.DefaultWorkspaceUser,
		DevcontainerFilePath: create.DEVCONTAINER_FILEPATH,
	}

	*config.CreateWorkspaceDtos, err = GetWorkspacesCreationDataFromPrompt(ctx, WorkspacesDataPromptParams{
		UserGitProviders: gitProviders,
		WorkspaceConfigs: workspaceConfigs,
		Manual:           *config.WorkspaceConfigurationFlags.Manual,
		MultiWorkspace:   config.MultiWorkspaceFlag,
		BlankWorkspace:   config.BlankFlag,
		ApiClient:        config.ApiClient,
		Defaults:         workspaceDefaults,
	})

	if err != nil {
		return err
	}

	generateWorkspaceIds(config.CreateWorkspaceDtos)
	setInitialWorkspaceNames(config.CreateWorkspaceDtos, *config.ExistingWorkspaces)

	submissionFormConfig := create.SubmissionFormParams{
		ChosenName:    &config.TargetName,
		WorkspaceList: config.CreateWorkspaceDtos,
		NameLabel:     config.TargetName,
		Defaults:      workspaceDefaults,
		ExistingWorkspaceNames: util.ArrayMap(*config.ExistingWorkspaces, func(w apiclient.WorkspaceDTO) string {
			return w.Name
		}),
	}

	err = create.RunSubmissionForm(submissionFormConfig)
	if err != nil {
		return err
	}

	return nil
}
