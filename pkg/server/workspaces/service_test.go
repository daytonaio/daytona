// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	t_targets "github.com/daytonaio/daytona/internal/testing/server/targets"
	"github.com/daytonaio/daytona/internal/testing/server/targets/mocks"
	t_workspaces "github.com/daytonaio/daytona/internal/testing/server/workspaces"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/apikey"
	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/server/workspaces"
	"github.com/daytonaio/daytona/pkg/server/workspaces/dto"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const serverApiUrl = "http://localhost:3986"
const serverUrl = "http://localhost:3987"
const serverVersion = "0.0.0-test"
const defaultWorkspaceUser = "daytona"
const defaultWorkspaceImage = "daytonaio/workspace-project:latest"

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

var tg = &target.Target{
	Id:   "123",
	Name: "test",
	ProviderInfo: target.ProviderInfo{
		Name:    "test-provider",
		Version: "test",
	},
	Options: "test-options",
}

var createWorkspaceDTO = dto.CreateWorkspaceDTO{
	Id:                  "123",
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
	Image:    util.Pointer(defaultWorkspaceImage),
	User:     util.Pointer(defaultWorkspaceUser),
	TargetId: tg.Id,
}

var workspaceInfo = workspace.WorkspaceInfo{
	Name:             createWorkspaceDTO.Name,
	Created:          "1 min ago",
	IsRunning:        true,
	ProviderMetadata: "provider-metadata-test",
	TargetId:         "123",
}

var ws = &workspace.Workspace{
	Id:                  "123",
	Name:                "workspace1",
	ApiKey:              createWorkspaceDTO.Name,
	GitProviderConfigId: &gitProviderConfig.Id,
	Repository: &gitprovider.GitRepository{
		Id:     "123",
		Url:    "https://github.com/daytonaio/daytona",
		Name:   "daytona",
		Branch: "main",
		Sha:    "sha1",
	},
	Image:    defaultWorkspaceImage,
	User:     defaultWorkspaceUser,
	TargetId: tg.Id,
}

func TestTargetService(t *testing.T) {
	ws.EnvVars = workspace.GetWorkspaceEnvVars(ws, workspace.WorkspaceEnvVarParams{
		ApiUrl:    serverApiUrl,
		ServerUrl: serverUrl,
		ClientId:  "test-client-id",
	}, false)

	ctx := context.Background()
	ctx = context.WithValue(ctx, telemetry.CLIENT_ID_CONTEXT_KEY, "test-client-id")

	targetStore := t_targets.NewInMemoryTargetStore()
	err := targetStore.Save(tg)
	require.Nil(t, err)

	workspaceStore := t_workspaces.NewInMemoryWorkspaceStore()

	containerRegistryService := mocks.NewMockContainerRegistryService()

	apiKeyService := mocks.NewMockApiKeyService()
	gitProviderService := mocks.NewMockGitProviderService()
	mockProvisioner := mocks.NewMockProvisioner()

	tgLogsDir := t.TempDir()
	buildLogsDir := t.TempDir()

	service := workspaces.NewWorkspaceService(workspaces.WorkspaceServiceConfig{
		TargetStore:              targetStore,
		WorkspaceStore:           workspaceStore,
		ServerApiUrl:             serverApiUrl,
		ServerUrl:                serverUrl,
		ContainerRegistryService: containerRegistryService,
		DefaultWorkspaceImage:    defaultWorkspaceImage,
		DefaultWorkspaceUser:     defaultWorkspaceUser,
		ApiKeyService:            apiKeyService,
		Provisioner:              mockProvisioner,
		LoggerFactory:            logs.NewLoggerFactory(&tgLogsDir, &buildLogsDir),
		GitProviderService:       gitProviderService,
	})

	t.Run("CreateWorkspace", func(t *testing.T) {
		var containerRegistry *containerregistry.ContainerRegistry

		containerRegistryService.On("FindByImageName", defaultWorkspaceImage).Return(containerRegistry, containerregistry.ErrContainerRegistryNotFound)

		gitProviderService.On("GetLastCommitSha", createWorkspaceDTO.Source.Repository).Return("123", nil)

		apiKeyService.On("Generate", apikey.ApiKeyTypeWorkspace, fmt.Sprintf("ws-%s", createWorkspaceDTO.Id)).Return(createWorkspaceDTO.Name, nil)

		ws := &workspace.Workspace{
			Name:                createWorkspaceDTO.Name,
			Image:               *createWorkspaceDTO.Image,
			User:                *createWorkspaceDTO.User,
			BuildConfig:         createWorkspaceDTO.BuildConfig,
			Repository:          createWorkspaceDTO.Source.Repository,
			ApiKey:              createWorkspaceDTO.Name,
			GitProviderConfigId: createWorkspaceDTO.GitProviderConfigId,
			TargetId:            createWorkspaceDTO.Id,
		}

		ws.EnvVars = workspace.GetWorkspaceEnvVars(ws, workspace.WorkspaceEnvVarParams{
			ApiUrl:        serverApiUrl,
			ServerUrl:     serverUrl,
			ServerVersion: serverVersion,
			ClientId:      "test",
		}, false)

		mockProvisioner.On("CreateWorkspace", provisioner.WorkspaceParams{
			Workspace:                     ws,
			Target:                        tg,
			ContainerRegistry:             containerRegistry,
			GitProviderConfig:             &gitProviderConfig,
			BuilderImage:                  defaultWorkspaceImage,
			BuilderImageContainerRegistry: containerRegistry,
		}).Return(nil)
		mockProvisioner.On("StartWorkspace", provisioner.WorkspaceParams{
			Workspace:                     ws,
			Target:                        tg,
			ContainerRegistry:             containerRegistry,
			GitProviderConfig:             &gitProviderConfig,
			BuilderImage:                  defaultWorkspaceImage,
			BuilderImageContainerRegistry: containerRegistry,
		}).Return(nil)

		gitProviderService.On("GetConfig", "github").Return(&gitProviderConfig, nil)

		workspace, err := service.CreateWorkspace(ctx, createWorkspaceDTO)

		require.Nil(t, err)
		require.NotNil(t, workspace)

		workspaceEquals(t, ws, &workspace.Workspace, defaultWorkspaceImage)

		ws.EnvVars = nil
	})

	t.Run("CreateWorkspace fails when workspace already exists", func(t *testing.T) {
		_, err := service.CreateWorkspace(ctx, createWorkspaceDTO)
		require.NotNil(t, err)
		require.Equal(t, workspaces.ErrWorkspaceAlreadyExists, err)
	})

	t.Run("CreateWorkspace fails name validation", func(t *testing.T) {
		invalidWorkspaceRequest := createWorkspaceDTO
		invalidWorkspaceRequest.Name = "invalid name"

		_, err := service.CreateWorkspace(ctx, invalidWorkspaceRequest)
		require.NotNil(t, err)
		require.Equal(t, workspaces.ErrInvalidWorkspaceName, err)
	})

	t.Run("GetWorkspace", func(t *testing.T) {
		mockProvisioner.On("GetWorkspaceInfo", ws, tg).Return(&workspaceInfo, nil)

		w, err := service.GetWorkspace(ctx, ws.Id, true)

		require.Nil(t, err)
		require.NotNil(t, w)

		workspaceDtoEquals(t, createWorkspaceDTO, *w, workspaceInfo, defaultWorkspaceImage, true)
	})

	t.Run("GetWorkspace fails when workspace not found", func(t *testing.T) {
		_, err := service.GetWorkspace(ctx, "invalid-id", true)
		require.NotNil(t, err)
		require.Equal(t, workspaces.ErrWorkspaceNotFound, err)
	})

	t.Run("ListWorkspaces", func(t *testing.T) {
		verbose := false

		workspaces, err := service.ListWorkspaces(ctx, verbose)

		require.Nil(t, err)
		require.Len(t, workspaces, 1)

		workspaceDtoEquals(t, createWorkspaceDTO, workspaces[0], workspaceInfo, defaultWorkspaceImage, verbose)
	})

	t.Run("ListWorkspaces - verbose", func(t *testing.T) {
		t.Skip("Need to figure out how to test the ListWorkspaces goroutine")

		verbose := true
		mockProvisioner.On("GetWorkspaceInfo", ws, tg).Return(&workspaceInfo, nil)

		workspaces, err := service.ListWorkspaces(ctx, verbose)

		require.Nil(t, err)
		require.Len(t, workspaces, 1)

		workspaceDtoEquals(t, createWorkspaceDTO, workspaces[0], workspaceInfo, defaultWorkspaceImage, verbose)
	})

	t.Run("StartWorkspace", func(t *testing.T) {
		mockProvisioner.On("StartWorkspace", mock.Anything, tg).Return(nil)

		err := service.StartWorkspace(ctx, createWorkspaceDTO.Id)

		require.Nil(t, err)
	})

	t.Run("StopWorkspace", func(t *testing.T) {
		mockProvisioner.On("StopWorkspace", mock.Anything, tg).Return(nil)

		err := service.StopWorkspace(ctx, createWorkspaceDTO.Id)

		require.Nil(t, err)
	})

	t.Run("RemoveWorkspace", func(t *testing.T) {
		mockProvisioner.On("DestroyWorkspace", mock.Anything, tg).Return(nil)
		apiKeyService.On("Revoke", mock.Anything).Return(nil)

		err := service.RemoveWorkspace(ctx, createWorkspaceDTO.Id)

		require.Nil(t, err)

		_, err = service.GetWorkspace(ctx, createWorkspaceDTO.Id, true)
		require.Equal(t, workspaces.ErrWorkspaceNotFound, err)
	})

	t.Run("ForceRemoveWorkspace", func(t *testing.T) {
		err := workspaceStore.Save(ws)
		require.Nil(t, err)

		mockProvisioner.On("DestroyWorkspace", mock.Anything, tg).Return(nil)
		apiKeyService.On("Revoke", mock.Anything).Return(nil)

		err = service.ForceRemoveWorkspace(ctx, createWorkspaceDTO.Id)

		require.Nil(t, err)

		_, err = service.GetWorkspace(ctx, createWorkspaceDTO.Id, true)
		require.Equal(t, workspaces.ErrWorkspaceNotFound, err)
	})

	t.Run("SetWorkspaceState", func(t *testing.T) {
		err := workspaceStore.Save(ws)
		require.Nil(t, err)

		updatedAt := time.Now().Format(time.RFC1123)
		res, err := service.SetWorkspaceState(createWorkspaceDTO.Id, &workspace.WorkspaceState{
			UpdatedAt: updatedAt,
			Uptime:    10,
			GitStatus: &workspace.GitStatus{
				CurrentBranch: "main",
			},
		})
		require.Nil(t, err)

		require.Nil(t, err)
		require.Equal(t, "main", res.State.GitStatus.CurrentBranch)
	})

	t.Cleanup(func() {
		apiKeyService.AssertExpectations(t)
		mockProvisioner.AssertExpectations(t)
	})
}

func workspaceEquals(t *testing.T, ws1, ws2 *workspace.Workspace, workspaceImage string) {
	t.Helper()

	require.Equal(t, ws1.Id, ws2.Id)
	require.Equal(t, ws1.Name, ws2.Name)
	require.Equal(t, ws1.TargetId, ws2.TargetId)
	require.Equal(t, ws1.Image, ws2.Image)
	require.Equal(t, ws1.User, ws2.User)
	require.Equal(t, ws1.ApiKey, ws2.ApiKey)
	require.Equal(t, ws1.Repository.Id, ws2.Repository.Id)
	require.Equal(t, ws1.Repository.Url, ws2.Repository.Url)
	require.Equal(t, ws1.Repository.Name, ws2.Repository.Name)
}

func workspaceDtoEquals(t *testing.T, req dto.CreateWorkspaceDTO, workspace dto.WorkspaceDTO, workspaceInfo workspace.WorkspaceInfo, workspaceImage string, verbose bool) {
	t.Helper()

	require.Equal(t, req.Id, workspace.Id)
	require.Equal(t, req.Name, workspace.Name)

	if verbose {
		require.Equal(t, workspace.Info.Name, workspaceInfo.Name)
		require.Equal(t, workspace.Info.ProviderMetadata, workspaceInfo.ProviderMetadata)
	} else {
		require.Nil(t, workspace.Info)
	}

	require.Equal(t, req.Name, workspace.Name)
	require.Equal(t, req.Source.Repository.Id, workspace.Repository.Id)
	require.Equal(t, req.Source.Repository.Url, workspace.Repository.Url)
	require.Equal(t, req.Source.Repository.Name, workspace.Repository.Name)
	require.Equal(t, workspace.ApiKey, workspace.Name)
	require.Equal(t, workspace.Image, workspaceImage)

	if verbose {
		require.Equal(t, workspace.Info.Name, workspaceInfo.Name)
		require.Equal(t, workspace.Info.Created, workspaceInfo.Created)
		require.Equal(t, workspace.Info.IsRunning, workspaceInfo.IsRunning)
		require.Equal(t, workspace.Info.ProviderMetadata, workspaceInfo.ProviderMetadata)
	}
}
