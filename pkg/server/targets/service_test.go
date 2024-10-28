// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	t_targetconfigs "github.com/daytonaio/daytona/internal/testing/provider/targetconfigs"
	t_targets "github.com/daytonaio/daytona/internal/testing/server/targets"
	"github.com/daytonaio/daytona/internal/testing/server/targets/mocks"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/apikey"
	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/server/targets"
	"github.com/daytonaio/daytona/pkg/server/targets/dto"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/target/workspace"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const serverApiUrl = "http://localhost:3986"
const serverUrl = "http://localhost:3987"
const serverVersion = "0.0.0-test"
const defaultWorkspaceUser = "daytona"
const defaultWorkspaceImage = "daytonaio/workspace-project:latest"

var targetConfig = provider.TargetConfig{
	Name: "test-target-config",
	ProviderInfo: provider.ProviderInfo{
		Name:    "test-provider",
		Version: "test",
	},
	Options: "test-options",
}
var gitProviderConfigId = "github"

var baseApiUrl = "https://api.github.com"

var gitProviderConfig = gitprovider.GitProviderConfig{
	Id:         "github",
	ProviderId: gitProviderConfigId,
	Alias:      "test-alias",
	Username:   "test-username",
	Token:      "test-token",
	BaseApiUrl: &baseApiUrl,
}

var createTargetDTO = dto.CreateTargetDTO{
	Name:         "test",
	Id:           "test",
	TargetConfig: targetConfig.Name,
	Workspaces: []dto.CreateWorkspaceDTO{
		{
			Name:                "workspace1",
			GitProviderConfigId: &gitProviderConfig.Id,
			Source: dto.CreateWorkspaceSourceDTO{
				Repository: &gitprovider.GitRepository{
					Id:     "123",
					Url:    "https://github.com/daytonaio/daytona",
					Name:   "daytona",
					Branch: "main",
					Sha:    "sha1",
				},
			},
			Image: util.Pointer(defaultWorkspaceImage),
			User:  util.Pointer(defaultWorkspaceUser),
		},
	},
}

var targetInfo = target.TargetInfo{
	Name:             createTargetDTO.Name,
	ProviderMetadata: "provider-metadata-test",
	Workspaces: []*workspace.WorkspaceInfo{
		{
			Name:             createTargetDTO.Workspaces[0].Name,
			Created:          "1 min ago",
			IsRunning:        true,
			ProviderMetadata: "provider-metadata-test",
			TargetId:         createTargetDTO.Id,
		},
	},
}

func TestTargetService(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, telemetry.CLIENT_ID_CONTEXT_KEY, "test")

	targetStore := t_targets.NewInMemoryTargetStore()

	containerRegistryService := mocks.NewMockContainerRegistryService()

	workspaceConfigService := mocks.NewMockWorkspaceConfigService()

	targetConfigStore := t_targetconfigs.NewInMemoryTargetConfigStore()
	err := targetConfigStore.Save(&targetConfig)
	require.Nil(t, err)

	apiKeyService := mocks.NewMockApiKeyService()
	gitProviderService := mocks.NewMockGitProviderService()
	mockProvisioner := mocks.NewMockProvisioner()

	tgLogsDir := t.TempDir()
	buildLogsDir := t.TempDir()

	service := targets.NewTargetService(targets.TargetServiceConfig{
		TargetStore:              targetStore,
		TargetConfigStore:        targetConfigStore,
		ServerApiUrl:             serverApiUrl,
		ServerUrl:                serverUrl,
		ServerVersion:            serverVersion,
		ContainerRegistryService: containerRegistryService,
		WorkspaceConfigService:   workspaceConfigService,
		DefaultWorkspaceImage:    defaultWorkspaceImage,
		DefaultWorkspaceUser:     defaultWorkspaceUser,
		ApiKeyService:            apiKeyService,
		BuilderImage:             defaultWorkspaceImage,
		Provisioner:              mockProvisioner,
		LoggerFactory:            logs.NewLoggerFactory(&tgLogsDir, &buildLogsDir),
		GitProviderService:       gitProviderService,
	})

	t.Run("CreateTarget", func(t *testing.T) {
		var containerRegistry *containerregistry.ContainerRegistry

		containerRegistryService.On("FindByImageName", defaultWorkspaceImage).Return(containerRegistry, containerregistry.ErrContainerRegistryNotFound)

		mockProvisioner.On("CreateTarget", mock.Anything, &targetConfig).Return(nil)
		mockProvisioner.On("StartTarget", mock.Anything, &targetConfig).Return(nil)

		apiKeyService.On("Generate", apikey.ApiKeyTypeTarget, createTargetDTO.Id).Return(createTargetDTO.Id, nil)
		gitProviderService.On("GetLastCommitSha", createTargetDTO.Workspaces[0].Source.Repository).Return("123", nil)

		for _, workspace := range createTargetDTO.Workspaces {
			apiKeyService.On("Generate", apikey.ApiKeyTypeWorkspace, fmt.Sprintf("%s/%s", createTargetDTO.Id, workspace.Name)).Return(workspace.Name, nil)
		}

		ws := &workspace.Workspace{
			Name:                createTargetDTO.Workspaces[0].Name,
			Image:               *createTargetDTO.Workspaces[0].Image,
			User:                *createTargetDTO.Workspaces[0].User,
			BuildConfig:         createTargetDTO.Workspaces[0].BuildConfig,
			Repository:          createTargetDTO.Workspaces[0].Source.Repository,
			ApiKey:              createTargetDTO.Workspaces[0].Name,
			GitProviderConfigId: createTargetDTO.Workspaces[0].GitProviderConfigId,
			TargetId:            createTargetDTO.Id,
			TargetConfig:        createTargetDTO.TargetConfig,
		}

		ws.EnvVars = workspace.GetWorkspaceEnvVars(ws, workspace.WorkspaceEnvVarParams{
			ApiUrl:        serverApiUrl,
			ServerUrl:     serverUrl,
			ServerVersion: serverVersion,
			ClientId:      "test",
		}, false)

		mockProvisioner.On("CreateWorkspace", provisioner.WorkspaceParams{
			Workspace:                     ws,
			TargetConfig:                  &targetConfig,
			ContainerRegistry:             containerRegistry,
			GitProviderConfig:             &gitProviderConfig,
			BuilderImage:                  defaultWorkspaceImage,
			BuilderImageContainerRegistry: containerRegistry,
		}).Return(nil)
		mockProvisioner.On("StartWorkspace", provisioner.WorkspaceParams{
			Workspace:                     ws,
			TargetConfig:                  &targetConfig,
			ContainerRegistry:             containerRegistry,
			GitProviderConfig:             &gitProviderConfig,
			BuilderImage:                  defaultWorkspaceImage,
			BuilderImageContainerRegistry: containerRegistry,
		}).Return(nil)

		gitProviderService.On("GetConfig", "github").Return(&gitProviderConfig, nil)

		target, err := service.CreateTarget(ctx, createTargetDTO)

		require.Nil(t, err)
		require.NotNil(t, target)

		targetEquals(t, createTargetDTO, target, defaultWorkspaceImage)
	})

	t.Run("CreateTarget fails when target already exists", func(t *testing.T) {
		_, err := service.CreateTarget(ctx, createTargetDTO)
		require.NotNil(t, err)
		require.Equal(t, targets.ErrTargetAlreadyExists, err)
	})

	t.Run("CreateTarget fails name validation", func(t *testing.T) {
		invalidTargetRequest := createTargetDTO
		invalidTargetRequest.Name = "invalid name"

		_, err := service.CreateTarget(ctx, invalidTargetRequest)
		require.NotNil(t, err)
		require.Equal(t, targets.ErrInvalidTargetName, err)
	})

	t.Run("GetTarget", func(t *testing.T) {
		mockProvisioner.On("GetTargetInfo", mock.Anything, mock.Anything, &targetConfig).Return(&targetInfo, nil)

		target, err := service.GetTarget(ctx, createTargetDTO.Id, true)

		require.Nil(t, err)
		require.NotNil(t, target)

		targetDtoEquals(t, createTargetDTO, *target, targetInfo, defaultWorkspaceImage, true)
	})

	t.Run("GetTarget fails when target not found", func(t *testing.T) {
		_, err := service.GetTarget(ctx, "invalid-id", true)
		require.NotNil(t, err)
		require.Equal(t, targets.ErrTargetNotFound, err)
	})

	t.Run("ListTargets", func(t *testing.T) {
		verbose := false
		mockProvisioner.On("GetTargetInfo", mock.Anything, mock.Anything, &targetConfig).Return(&targetInfo, nil)

		targets, err := service.ListTargets(ctx, verbose)

		require.Nil(t, err)
		require.Len(t, targets, 1)

		target := targets[0]

		targetDtoEquals(t, createTargetDTO, target, targetInfo, defaultWorkspaceImage, verbose)
	})

	t.Run("ListTargets - verbose", func(t *testing.T) {
		verbose := true
		mockProvisioner.On("GetTargetInfo", mock.Anything, mock.Anything, &targetConfig).Return(&targetInfo, nil)

		targets, err := service.ListTargets(ctx, verbose)

		require.Nil(t, err)
		require.Len(t, targets, 1)

		target := targets[0]

		targetDtoEquals(t, createTargetDTO, target, targetInfo, defaultWorkspaceImage, verbose)
	})

	t.Run("StartTarget", func(t *testing.T) {
		mockProvisioner.On("StartTarget", mock.Anything, &targetConfig).Return(nil)
		mockProvisioner.On("StartWorkspace", mock.Anything, &targetConfig).Return(nil)

		err := service.StartTarget(ctx, createTargetDTO.Id)

		require.Nil(t, err)
	})

	t.Run("StartWorkspace", func(t *testing.T) {
		mockProvisioner.On("StartTarget", mock.Anything, &targetConfig).Return(nil)
		mockProvisioner.On("StartWorkspace", mock.Anything, &targetConfig).Return(nil)

		err := service.StartWorkspace(ctx, createTargetDTO.Id, createTargetDTO.Workspaces[0].Name)

		require.Nil(t, err)
	})

	t.Run("StopTarget", func(t *testing.T) {
		mockProvisioner.On("StopTarget", mock.Anything, &targetConfig).Return(nil)
		mockProvisioner.On("StopWorkspace", mock.Anything, &targetConfig).Return(nil)

		err := service.StopTarget(ctx, createTargetDTO.Id)

		require.Nil(t, err)
	})

	t.Run("StopWorkspace", func(t *testing.T) {
		mockProvisioner.On("StopTarget", mock.Anything, &targetConfig).Return(nil)
		mockProvisioner.On("StopWorkspace", mock.Anything, &targetConfig).Return(nil)

		err := service.StopWorkspace(ctx, createTargetDTO.Id, createTargetDTO.Workspaces[0].Name)

		require.Nil(t, err)
	})

	t.Run("RemoveTarget", func(t *testing.T) {
		mockProvisioner.On("DestroyTarget", mock.Anything, &targetConfig).Return(nil)
		mockProvisioner.On("DestroyWorkspace", mock.Anything, &targetConfig).Return(nil)
		apiKeyService.On("Revoke", mock.Anything).Return(nil)

		err := service.RemoveTarget(ctx, createTargetDTO.Id)

		require.Nil(t, err)

		_, err = service.GetTarget(ctx, createTargetDTO.Id, true)
		require.Equal(t, targets.ErrTargetNotFound, err)
	})

	t.Run("ForceRemoveTarget", func(t *testing.T) {
		err := targetStore.Save(&target.Target{Id: createTargetDTO.Id, TargetConfig: targetConfig.Name})
		require.Nil(t, err)

		mockProvisioner.On("DestroyTarget", mock.Anything, &targetConfig).Return(nil)
		mockProvisioner.On("DestroyWorkspace", mock.Anything, &targetConfig).Return(nil)
		apiKeyService.On("Revoke", mock.Anything).Return(nil)

		err = service.ForceRemoveTarget(ctx, createTargetDTO.Id)

		require.Nil(t, err)

		_, err = service.GetTarget(ctx, createTargetDTO.Id, true)
		require.Equal(t, targets.ErrTargetNotFound, err)
	})

	t.Run("SetWorkspaceState", func(t *testing.T) {
		tg, err := service.CreateTarget(ctx, createTargetDTO)
		require.Nil(t, err)

		workspaceName := tg.Workspaces[0].Name
		updatedAt := time.Now().Format(time.RFC1123)
		res, err := service.SetWorkspaceState(tg.Id, workspaceName, &workspace.WorkspaceState{
			UpdatedAt: updatedAt,
			Uptime:    10,
			GitStatus: &workspace.GitStatus{
				CurrentBranch: "main",
			},
		})
		require.Nil(t, err)

		workspace, err := res.GetWorkspace(workspaceName)
		require.Nil(t, err)
		require.Equal(t, "main", workspace.State.GitStatus.CurrentBranch)
	})

	t.Cleanup(func() {
		apiKeyService.AssertExpectations(t)
		mockProvisioner.AssertExpectations(t)
	})
}

func targetEquals(t *testing.T, req dto.CreateTargetDTO, target *target.Target, workspaceImage string) {
	t.Helper()

	require.Equal(t, req.Id, target.Id)
	require.Equal(t, req.Name, target.Name)
	require.Equal(t, req.TargetConfig, target.TargetConfig)

	for i, workspace := range target.Workspaces {
		require.Equal(t, req.Workspaces[i].Name, workspace.Name)
		require.Equal(t, req.Workspaces[i].Source.Repository.Id, workspace.Repository.Id)
		require.Equal(t, req.Workspaces[i].Source.Repository.Url, workspace.Repository.Url)
		require.Equal(t, req.Workspaces[i].Source.Repository.Name, workspace.Repository.Name)
		require.Equal(t, workspace.ApiKey, workspace.Name)
		require.Equal(t, workspace.TargetConfig, req.TargetConfig)
		require.Equal(t, workspace.Image, workspaceImage)
	}
}

func targetDtoEquals(t *testing.T, req dto.CreateTargetDTO, target dto.TargetDTO, targetInfo target.TargetInfo, workspaceImage string, verbose bool) {
	t.Helper()

	require.Equal(t, req.Id, target.Id)
	require.Equal(t, req.Name, target.Name)
	require.Equal(t, req.TargetConfig, target.TargetConfig)

	if verbose {
		require.Equal(t, target.Info.Name, targetInfo.Name)
		require.Equal(t, target.Info.ProviderMetadata, targetInfo.ProviderMetadata)
	} else {
		require.Nil(t, target.Info)
	}

	for i, workspace := range target.Workspaces {
		require.Equal(t, req.Workspaces[i].Name, workspace.Name)
		require.Equal(t, req.Workspaces[i].Source.Repository.Id, workspace.Repository.Id)
		require.Equal(t, req.Workspaces[i].Source.Repository.Url, workspace.Repository.Url)
		require.Equal(t, req.Workspaces[i].Source.Repository.Name, workspace.Repository.Name)
		require.Equal(t, workspace.ApiKey, workspace.Name)
		require.Equal(t, workspace.TargetConfig, req.TargetConfig)
		require.Equal(t, workspace.Image, workspaceImage)

		if verbose {
			require.Equal(t, target.Info.Workspaces[i].Name, targetInfo.Workspaces[i].Name)
			require.Equal(t, target.Info.Workspaces[i].Created, targetInfo.Workspaces[i].Created)
			require.Equal(t, target.Info.Workspaces[i].IsRunning, targetInfo.Workspaces[i].IsRunning)
			require.Equal(t, target.Info.Workspaces[i].ProviderMetadata, targetInfo.Workspaces[i].ProviderMetadata)
		}
	}
}
