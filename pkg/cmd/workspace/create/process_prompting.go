// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"context"
	"errors"

	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	cmd_common "github.com/daytonaio/daytona/pkg/cmd/common"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/create"
)

type ProcessPromptingParams struct {
	ApiClient                   *apiclient.APIClient
	CreateWorkspaceDtos         *[]apiclient.CreateWorkspaceDTO
	ExistingWorkspaces          *[]apiclient.WorkspaceDTO
	WorkspaceConfigurationFlags cmd_common.WorkspaceConfigurationFlags
	MultiWorkspaceFlag          bool
	BlankFlag                   bool
	TargetName                  string
}

func ProcessPrompting(ctx context.Context, params ProcessPromptingParams) error {
	if cmd_common.CheckAnyWorkspaceConfigurationFlagSet(params.WorkspaceConfigurationFlags) || (params.WorkspaceConfigurationFlags.Branches != nil && len(*params.WorkspaceConfigurationFlags.Branches) > 0) {
		return errors.New("please provide the repository URL in order to set up custom workspace details through the CLI")
	}

	gitProviders, res, err := params.ApiClient.GitProviderAPI.ListGitProviders(ctx).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}

	workspaceConfigs, res, err := params.ApiClient.WorkspaceConfigAPI.ListWorkspaceConfigs(ctx).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}

	apiServerConfig, res, err := params.ApiClient.ServerAPI.GetConfig(ctx).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}

	workspaceDefaults := &views_util.WorkspaceConfigDefaults{
		BuildChoice:          views_util.AUTOMATIC,
		Image:                &apiServerConfig.DefaultWorkspaceImage,
		ImageUser:            &apiServerConfig.DefaultWorkspaceUser,
		DevcontainerFilePath: create.DEVCONTAINER_FILEPATH,
	}

	*params.CreateWorkspaceDtos, err = GetWorkspacesCreationDataFromPrompt(ctx, WorkspacesDataPromptParams{
		UserGitProviders: gitProviders,
		WorkspaceConfigs: workspaceConfigs,
		Manual:           *params.WorkspaceConfigurationFlags.Manual,
		MultiWorkspace:   params.MultiWorkspaceFlag,
		BlankWorkspace:   params.BlankFlag,
		ApiClient:        params.ApiClient,
		Defaults:         workspaceDefaults,
	})

	if err != nil {
		return err
	}

	generateWorkspaceIds(params.CreateWorkspaceDtos)
	setInitialWorkspaceNames(params.CreateWorkspaceDtos, *params.ExistingWorkspaces)

	submissionFormConfig := create.SubmissionFormParams{
		ChosenName:    &params.TargetName,
		WorkspaceList: params.CreateWorkspaceDtos,
		NameLabel:     params.TargetName,
		Defaults:      workspaceDefaults,
		ExistingWorkspaceNames: util.ArrayMap(*params.ExistingWorkspaces, func(w apiclient.WorkspaceDTO) string {
			return w.Name
		}),
	}

	err = create.RunSubmissionForm(submissionFormConfig)
	if err != nil {
		return err
	}

	return nil
}
