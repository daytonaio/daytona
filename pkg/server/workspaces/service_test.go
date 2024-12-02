// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces_test

import (
	"context"
	"fmt"
	"testing"

	t_targets "github.com/daytonaio/daytona/internal/testing/server/targets"
	"github.com/daytonaio/daytona/internal/testing/server/targets/mocks"
	t_workspaces "github.com/daytonaio/daytona/internal/testing/server/workspaces"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/server/workspaces"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/daytonaio/daytona/pkg/telemetry"
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

var gitProviderConfig = models.GitProviderConfig{
	Id:         "github",
	ProviderId: gitProviderConfigId,
	Alias:      "test-alias",
	Username:   "test-username",
	Token:      "test-token",
	BaseApiUrl: &baseApiUrl,
}

var tc = &models.TargetConfig{
	Name: "tc-test",
	ProviderInfo: models.ProviderInfo{
		Name:    "test-provider",
		Version: "test",
	},
	Options: "test-options",
	Deleted: false,
}

var tg = &models.Target{
	Id:             "123",
	Name:           "test",
	TargetConfigId: tc.Id,
	TargetConfig:   *tc,
}

var createWorkspaceDTO = services.CreateWorkspaceDTO{
	Id:                  "123",
	Name:                "workspace1",
	GitProviderConfigId: &gitProviderConfig.Id,
	Source: services.CreateWorkspaceSourceDTO{
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

var workspaceInfo = models.WorkspaceInfo{
	Name:             createWorkspaceDTO.Name,
	Created:          "1 min ago",
	IsRunning:        true,
	ProviderMetadata: "provider-metadata-test",
	TargetId:         "123",
}

var ws = &models.Workspace{
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
	ws.EnvVars = workspaces.GetWorkspaceEnvVars(ws, workspaces.WorkspaceEnvVarParams{
		ApiUrl:    serverApiUrl,
		ServerUrl: serverUrl,
		ClientId:  "test-client-id",
	})

	ctx := context.Background()
	ctx = context.WithValue(ctx, telemetry.CLIENT_ID_CONTEXT_KEY, "test-client-id")

	targetStore := t_targets.NewInMemoryTargetStore()
	err := targetStore.Save(tg)
	require.Nil(t, err)

	workspaceStore := t_workspaces.NewInMemoryWorkspaceStore()

	apiKeyService := mocks.NewMockApiKeyService()
	gitProviderService := mocks.NewMockGitProviderService()
	mockProvisioner := mocks.NewMockProvisioner()

	tgLogsDir := t.TempDir()
	buildLogsDir := t.TempDir()

	service := workspaces.NewWorkspaceService(workspaces.WorkspaceServiceConfig{
		FindTarget: func(ctx context.Context, targetId string) (*models.Target, error) {
			t, err := targetStore.Find(&stores.TargetFilter{IdOrName: &targetId})
			if err != nil {
				return nil, err
			}
			return t, nil
		},
		FindContainerRegistry: func(ctx context.Context, image string, envVars map[string]string) *models.ContainerRegistry {
			return services.EnvironmentVariables(envVars).FindContainerRegistryByImageName(image)
		},
		FindCachedBuild: func(ctx context.Context, w *models.Workspace) (*models.CachedBuild, error) {
			return nil, nil
		},
		GenerateApiKey: func(ctx context.Context, name string) (string, error) {
			return apiKeyService.Generate(models.ApiKeyTypeWorkspace, name)
		},
		RevokeApiKey: func(ctx context.Context, name string) error {
			return apiKeyService.Revoke(name)
		},
		ListGitProviderConfigs: func(ctx context.Context, repoUrl string) ([]*models.GitProviderConfig, error) {
			return gitProviderService.ListConfigsForUrl(repoUrl)
		},
		FindGitProviderConfig: func(ctx context.Context, id string) (*models.GitProviderConfig, error) {
			return gitProviderService.GetConfig(id)
		},
		GetLastCommitSha: func(ctx context.Context, repo *gitprovider.GitRepository) (string, error) {
			return gitProviderService.GetLastCommitSha(repo)
		},
		TrackTelemetryEvent: func(event telemetry.ServerEvent, clientId string, props map[string]interface{}) error {
			return nil
		},
		WorkspaceStore:        workspaceStore,
		ServerApiUrl:          serverApiUrl,
		ServerUrl:             serverUrl,
		DefaultWorkspaceImage: defaultWorkspaceImage,
		DefaultWorkspaceUser:  defaultWorkspaceUser,
		Provisioner:           mockProvisioner,
		LoggerFactory:         logs.NewLoggerFactory(&tgLogsDir, &buildLogsDir),
	})

	t.Run("CreateWorkspace", func(t *testing.T) {
		var containerRegistry *models.ContainerRegistry

		gitProviderService.On("GetLastCommitSha", createWorkspaceDTO.Source.Repository).Return("123", nil)

		apiKeyService.On("Generate", models.ApiKeyTypeWorkspace, fmt.Sprintf("ws-%s", createWorkspaceDTO.Id)).Return(createWorkspaceDTO.Name, nil)

		ws := &models.Workspace{
			Name:                createWorkspaceDTO.Name,
			Image:               *createWorkspaceDTO.Image,
			User:                *createWorkspaceDTO.User,
			BuildConfig:         createWorkspaceDTO.BuildConfig,
			Repository:          createWorkspaceDTO.Source.Repository,
			ApiKey:              createWorkspaceDTO.Name,
			GitProviderConfigId: createWorkspaceDTO.GitProviderConfigId,
			TargetId:            createWorkspaceDTO.Id,
		}

		ws.EnvVars = workspaces.GetWorkspaceEnvVars(ws, workspaces.WorkspaceEnvVarParams{
			ApiUrl:        serverApiUrl,
			ServerUrl:     serverUrl,
			ServerVersion: serverVersion,
			ClientId:      "test",
		})

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

		workspaceEquals(t, &services.WorkspaceDTO{Workspace: *ws}, workspace)

		ws.EnvVars = nil
	})

	t.Run("CreateWorkspace fails when workspace already exists", func(t *testing.T) {
		_, err := service.CreateWorkspace(ctx, createWorkspaceDTO)
		require.NotNil(t, err)
		require.Equal(t, services.ErrWorkspaceAlreadyExists, err)
	})

	t.Run("CreateWorkspace fails name validation", func(t *testing.T) {
		invalidWorkspaceRequest := createWorkspaceDTO
		invalidWorkspaceRequest.Name = "invalid name"

		_, err := service.CreateWorkspace(ctx, invalidWorkspaceRequest)
		require.NotNil(t, err)
		require.Equal(t, services.ErrInvalidWorkspaceName, err)
	})

	t.Run("GetWorkspace", func(t *testing.T) {
		mockProvisioner.On("GetWorkspaceInfo", ws, tg).Return(&workspaceInfo, nil)

		w, err := service.GetWorkspace(ctx, ws.Id, services.WorkspaceRetrievalParams{Verbose: true})

		require.Nil(t, err)
		require.NotNil(t, w)

		workspaceDtoEquals(t, createWorkspaceDTO, *w, workspaceInfo, defaultWorkspaceImage, true)
	})

	t.Run("GetWorkspace fails when workspace not found", func(t *testing.T) {
		_, err := service.GetWorkspace(ctx, "invalid-id", services.WorkspaceRetrievalParams{Verbose: true})
		require.NotNil(t, err)
		require.Equal(t, stores.ErrWorkspaceNotFound, err)
	})

	t.Run("ListWorkspaces", func(t *testing.T) {
		verbose := false

		workspaces, err := service.ListWorkspaces(ctx, services.WorkspaceRetrievalParams{Verbose: verbose})

		require.Nil(t, err)
		require.Len(t, workspaces, 1)

		workspaceDtoEquals(t, createWorkspaceDTO, workspaces[0], workspaceInfo, defaultWorkspaceImage, verbose)
	})

	t.Run("ListWorkspaces - verbose", func(t *testing.T) {
		t.Skip("Need to figure out how to test the ListWorkspaces goroutine")

		verbose := true
		mockProvisioner.On("GetWorkspaceInfo", ws, tg).Return(&workspaceInfo, nil)

		workspaces, err := service.ListWorkspaces(ctx, services.WorkspaceRetrievalParams{Verbose: verbose})

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

		_, err = service.GetWorkspace(ctx, createWorkspaceDTO.Id, services.WorkspaceRetrievalParams{Verbose: true})
		require.Equal(t, stores.ErrWorkspaceNotFound, err)
	})

	t.Run("ForceRemoveWorkspace", func(t *testing.T) {
		err := workspaceStore.Save(ws)
		require.Nil(t, err)

		mockProvisioner.On("DestroyWorkspace", mock.Anything, tg).Return(nil)
		apiKeyService.On("Revoke", mock.Anything).Return(nil)

		err = service.ForceRemoveWorkspace(ctx, createWorkspaceDTO.Id)

		require.Nil(t, err)

		_, err = service.GetWorkspace(ctx, createWorkspaceDTO.Id, services.WorkspaceRetrievalParams{Verbose: true})
		require.Equal(t, stores.ErrWorkspaceNotFound, err)
	})

	t.Run("SetWorkspaceMetadata", func(t *testing.T) {
		err := workspaceStore.Save(ws)
		require.Nil(t, err)

		res, err := service.SetWorkspaceMetadata(createWorkspaceDTO.Id, &models.WorkspaceMetadata{
			Uptime: 10,
			GitStatus: &models.GitStatus{
				CurrentBranch: "main",
			},
		})
		require.Nil(t, err)

		require.Nil(t, err)
		require.Equal(t, "main", res.GitStatus.CurrentBranch)
	})

	t.Cleanup(func() {
		apiKeyService.AssertExpectations(t)
		mockProvisioner.AssertExpectations(t)
	})
}

func workspaceEquals(t *testing.T, ws1, ws2 *services.WorkspaceDTO) {
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

func workspaceDtoEquals(t *testing.T, req services.CreateWorkspaceDTO, workspace services.WorkspaceDTO, workspaceInfo models.WorkspaceInfo, workspaceImage string, verbose bool) {
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
