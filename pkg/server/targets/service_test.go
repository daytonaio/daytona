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
	"github.com/daytonaio/daytona/pkg/server/targets"
	"github.com/daytonaio/daytona/pkg/server/targets/dto"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/target/workspace"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const serverApiUrl = "http://localhost:3986"
const serverUrl = "http://localhost:3987"
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
	targetStore := t_targets.NewInMemoryTargetStore()

	containerRegistryService := mocks.NewMockContainerRegistryService()

	workspaceConfigService := mocks.NewMockWorkspaceConfigService()

	targetConfigStore := t_targetconfigs.NewInMemoryTargetConfigStore()
	err := targetConfigStore.Save(&targetConfig)
	require.Nil(t, err)

	apiKeyService := mocks.NewMockApiKeyService()
	gitProviderService := mocks.NewMockGitProviderService()
	provisioner := mocks.NewMockProvisioner()

	tgLogsDir := t.TempDir()
	buildLogsDir := t.TempDir()

	service := targets.NewTargetService(targets.TargetServiceConfig{
		TargetStore:              targetStore,
		TargetConfigStore:        targetConfigStore,
		ServerApiUrl:             serverApiUrl,
		ServerUrl:                serverUrl,
		ContainerRegistryService: containerRegistryService,
		WorkspaceConfigService:   workspaceConfigService,
		DefaultWorkspaceImage:    defaultWorkspaceImage,
		DefaultWorkspaceUser:     defaultWorkspaceUser,
		ApiKeyService:            apiKeyService,
		Provisioner:              provisioner,
		LoggerFactory:            logs.NewLoggerFactory(&tgLogsDir, &buildLogsDir),
		GitProviderService:       gitProviderService,
	})

	t.Run("CreateTarget", func(t *testing.T) {
		var containerRegistry *containerregistry.ContainerRegistry

		containerRegistryService.On("FindByImageName", defaultWorkspaceImage).Return(containerRegistry, containerregistry.ErrContainerRegistryNotFound)

		provisioner.On("CreateTarget", mock.Anything, &targetConfig).Return(nil)
		provisioner.On("StartTarget", mock.Anything, &targetConfig).Return(nil)

		apiKeyService.On("Generate", apikey.ApiKeyTypeTarget, createTargetDTO.Id).Return(createTargetDTO.Id, nil)
		gitProviderService.On("GetLastCommitSha", createTargetDTO.Workspaces[0].Source.Repository).Return("123", nil)

		for _, workspace := range createTargetDTO.Workspaces {
			apiKeyService.On("Generate", apikey.ApiKeyTypeWorkspace, fmt.Sprintf("%s/%s", createTargetDTO.Id, workspace.Name)).Return(workspace.Name, nil)
		}
		provisioner.On("CreateWorkspace", mock.Anything, &targetConfig, containerRegistry, &gitProviderConfig).Return(nil)
		provisioner.On("StartWorkspace", mock.Anything, &targetConfig).Return(nil)

		gitProviderService.On("GetConfig", "github").Return(&gitProviderConfig, nil)

		target, err := service.CreateTarget(context.TODO(), createTargetDTO)

		require.Nil(t, err)
		require.NotNil(t, target)

		targetEquals(t, createTargetDTO, target, defaultWorkspaceImage)
	})

	t.Run("CreateTarget fails when target already exists", func(t *testing.T) {
		_, err := service.CreateTarget(context.TODO(), createTargetDTO)
		require.NotNil(t, err)
		require.Equal(t, targets.ErrTargetAlreadyExists, err)
	})

	t.Run("CreateTarget fails name validation", func(t *testing.T) {
		invalidTargetRequest := createTargetDTO
		invalidTargetRequest.Name = "invalid name"

		_, err := service.CreateTarget(context.TODO(), invalidTargetRequest)
		require.NotNil(t, err)
		require.Equal(t, targets.ErrInvalidTargetName, err)
	})

	t.Run("GetTarget", func(t *testing.T) {
		provisioner.On("GetTargetInfo", mock.Anything, mock.Anything, &targetConfig).Return(&targetInfo, nil)

		target, err := service.GetTarget(context.TODO(), createTargetDTO.Id, true)

		require.Nil(t, err)
		require.NotNil(t, target)

		targetDtoEquals(t, createTargetDTO, *target, targetInfo, defaultWorkspaceImage, true)
	})

	t.Run("GetTarget fails when target not found", func(t *testing.T) {
		_, err := service.GetTarget(context.TODO(), "invalid-id", true)
		require.NotNil(t, err)
		require.Equal(t, targets.ErrTargetNotFound, err)
	})

	t.Run("ListTargets", func(t *testing.T) {
		verbose := false
		provisioner.On("GetTargetInfo", mock.Anything, mock.Anything, &targetConfig).Return(&targetInfo, nil)

		targets, err := service.ListTargets(context.TODO(), verbose)

		require.Nil(t, err)
		require.Len(t, targets, 1)

		target := targets[0]

		targetDtoEquals(t, createTargetDTO, target, targetInfo, defaultWorkspaceImage, verbose)
	})

	t.Run("ListTargets - verbose", func(t *testing.T) {
		verbose := true
		provisioner.On("GetTargetInfo", mock.Anything, mock.Anything, &targetConfig).Return(&targetInfo, nil)

		targets, err := service.ListTargets(context.TODO(), verbose)

		require.Nil(t, err)
		require.Len(t, targets, 1)

		target := targets[0]

		targetDtoEquals(t, createTargetDTO, target, targetInfo, defaultWorkspaceImage, verbose)
	})

	t.Run("StartTarget", func(t *testing.T) {
		provisioner.On("StartTarget", mock.Anything, &targetConfig).Return(nil)
		provisioner.On("StartWorkspace", mock.Anything, &targetConfig).Return(nil)

		err := service.StartTarget(context.TODO(), createTargetDTO.Id)

		require.Nil(t, err)
	})

	t.Run("StartWorkspace", func(t *testing.T) {
		provisioner.On("StartTarget", mock.Anything, &targetConfig).Return(nil)
		provisioner.On("StartWorkspace", mock.Anything, &targetConfig).Return(nil)

		err := service.StartWorkspace(context.TODO(), createTargetDTO.Id, createTargetDTO.Workspaces[0].Name)

		require.Nil(t, err)
	})

	t.Run("StopTarget", func(t *testing.T) {
		provisioner.On("StopTarget", mock.Anything, &targetConfig).Return(nil)
		provisioner.On("StopWorkspace", mock.Anything, &targetConfig).Return(nil)

		err := service.StopTarget(context.TODO(), createTargetDTO.Id)

		require.Nil(t, err)
	})

	t.Run("StopWorkspace", func(t *testing.T) {
		provisioner.On("StopTarget", mock.Anything, &targetConfig).Return(nil)
		provisioner.On("StopWorkspace", mock.Anything, &targetConfig).Return(nil)

		err := service.StopWorkspace(context.TODO(), createTargetDTO.Id, createTargetDTO.Workspaces[0].Name)

		require.Nil(t, err)
	})

	t.Run("RemoveTarget", func(t *testing.T) {
		provisioner.On("DestroyTarget", mock.Anything, &targetConfig).Return(nil)
		provisioner.On("DestroyWorkspace", mock.Anything, &targetConfig).Return(nil)
		apiKeyService.On("Revoke", mock.Anything).Return(nil)

		err := service.RemoveTarget(context.TODO(), createTargetDTO.Id)

		require.Nil(t, err)

		_, err = service.GetTarget(context.TODO(), createTargetDTO.Id, true)
		require.Equal(t, targets.ErrTargetNotFound, err)
	})

	t.Run("ForceRemoveTarget", func(t *testing.T) {
		var containerRegistry *containerregistry.ContainerRegistry

		provisioner.On("CreateTarget", mock.Anything, &targetConfig).Return(nil)
		provisioner.On("StartTarget", mock.Anything, &targetConfig).Return(nil)

		apiKeyService.On("Generate", apikey.ApiKeyTypeTarget, createTargetDTO.Id).Return(createTargetDTO.Id, nil)
		gitProviderService.On("GetLastCommitSha", createTargetDTO.Workspaces[0].Source.Repository).Return("123", nil)

		gitProviderService.On("ListConfigsForUrl", "https://github.com/daytonaio/daytona").Return([]*gitprovider.GitProviderConfig{&gitProviderConfig}, nil)

		for _, workspace := range createTargetDTO.Workspaces {
			apiKeyService.On("Generate", apikey.ApiKeyTypeWorkspace, fmt.Sprintf("%s/%s", createTargetDTO.Id, workspace.Name)).Return(workspace.Name, nil)
		}
		provisioner.On("CreateWorkspace", mock.Anything, &targetConfig, containerRegistry, &gitProviderConfig).Return(nil)
		provisioner.On("StartWorkspace", mock.Anything, &targetConfig).Return(nil)

		_, _ = service.CreateTarget(context.TODO(), createTargetDTO)

		provisioner.On("DestroyTarget", mock.Anything, &targetConfig).Return(nil)
		provisioner.On("DestroyWorkspace", mock.Anything, &targetConfig).Return(nil)
		apiKeyService.On("Revoke", mock.Anything).Return(nil)

		err = service.ForceRemoveTarget(context.TODO(), createTargetDTO.Id)

		require.Nil(t, err)

		_, err = service.GetTarget(context.TODO(), createTargetDTO.Id, true)
		require.Equal(t, targets.ErrTargetNotFound, err)
	})

	t.Run("SetWorkspaceState", func(t *testing.T) {
		tg, err := service.CreateTarget(context.TODO(), createTargetDTO)
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
		provisioner.AssertExpectations(t)
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
