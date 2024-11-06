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
	"github.com/daytonaio/daytona/pkg/server/workspaces"
	"github.com/daytonaio/daytona/pkg/server/workspaces/dto"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const serverApiUrl = "http://localhost:3986"
const serverUrl = "http://localhost:3987"
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

var ws *workspace.Workspace

func TestTargetService(t *testing.T) {
	targetStore := t_targets.NewInMemoryTargetStore()
	err := targetStore.Save(tg)
	require.Nil(t, err)

	workspaceStore := t_workspaces.NewInMemoryWorkspaceStore()

	containerRegistryService := mocks.NewMockContainerRegistryService()

	apiKeyService := mocks.NewMockApiKeyService()
	gitProviderService := mocks.NewMockGitProviderService()
	provisioner := mocks.NewMockProvisioner()

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
		Provisioner:              provisioner,
		LoggerFactory:            logs.NewLoggerFactory(&tgLogsDir, &buildLogsDir),
		GitProviderService:       gitProviderService,
	})

	t.Run("CreateWorkspace", func(t *testing.T) {
		var containerRegistry *containerregistry.ContainerRegistry

		containerRegistryService.On("FindByImageName", defaultWorkspaceImage).Return(containerRegistry, containerregistry.ErrContainerRegistryNotFound)

		gitProviderService.On("GetLastCommitSha", createWorkspaceDTO.Source.Repository).Return("123", nil)

		apiKeyService.On("Generate", apikey.ApiKeyTypeWorkspace, fmt.Sprintf("ws-%s", createWorkspaceDTO.Id)).Return(createWorkspaceDTO.Name, nil)
		provisioner.On("CreateWorkspace", mock.Anything, &tg, containerRegistry, &gitProviderConfig).Return(nil)
		provisioner.On("StartWorkspace", mock.Anything, &tg).Return(nil)

		gitProviderService.On("GetConfig", "github").Return(&gitProviderConfig, nil)

		workspace, err := service.CreateWorkspace(context.TODO(), createWorkspaceDTO)

		require.Nil(t, err)
		require.NotNil(t, workspace)

		ws = &workspace.Workspace
		ws.EnvVars = nil

		workspaceEquals(t, createWorkspaceDTO, &workspace.Workspace, defaultWorkspaceImage)
	})

	t.Run("CreateWorkspace fails when workspace already exists", func(t *testing.T) {
		_, err := service.CreateWorkspace(context.TODO(), createWorkspaceDTO)
		require.NotNil(t, err)
		require.Equal(t, workspaces.ErrWorkspaceAlreadyExists, err)
	})

	t.Run("CreateWorkspace fails name validation", func(t *testing.T) {
		invalidWorkspaceRequest := createWorkspaceDTO
		invalidWorkspaceRequest.Name = "invalid name"

		_, err := service.CreateWorkspace(context.TODO(), invalidWorkspaceRequest)
		require.NotNil(t, err)
		require.Equal(t, workspaces.ErrInvalidWorkspaceName, err)
	})

	t.Run("GetWorkspace", func(t *testing.T) {
		provisioner.On("GetWorkspaceInfo", mock.Anything, ws, &tg).Return(&workspaceInfo, nil)

		w, err := service.GetWorkspace(context.TODO(), ws.Id, true)

		require.Nil(t, err)
		require.NotNil(t, w)

		workspaceDtoEquals(t, createWorkspaceDTO, *w, workspaceInfo, defaultWorkspaceImage, true)
	})

	t.Run("GetWorkspace fails when workspace not found", func(t *testing.T) {
		_, err := service.GetWorkspace(context.TODO(), "invalid-id", true)
		require.NotNil(t, err)
		require.Equal(t, workspaces.ErrWorkspaceNotFound, err)
	})

	t.Run("ListWorkspaces", func(t *testing.T) {
		verbose := false

		workspaces, err := service.ListWorkspaces(context.TODO(), verbose)

		require.Nil(t, err)
		require.Len(t, workspaces, 1)

		workspaceDtoEquals(t, createWorkspaceDTO, workspaces[0], workspaceInfo, defaultWorkspaceImage, verbose)
	})

	t.Run("ListWorkspaces - verbose", func(t *testing.T) {
		t.Skip("Need to figure out how to test the ListWorkspaces goroutine")

		verbose := true
		provisioner.On("GetWorkspaceInfo", mock.Anything, ws, &tg).Return(&workspaceInfo, nil)

		workspaces, err := service.ListWorkspaces(context.TODO(), verbose)

		require.Nil(t, err)
		require.Len(t, workspaces, 1)

		workspaceDtoEquals(t, createWorkspaceDTO, workspaces[0], workspaceInfo, defaultWorkspaceImage, verbose)
	})

	t.Run("StartWorkspace", func(t *testing.T) {
		provisioner.On("StartWorkspace", mock.Anything, &tg).Return(nil)

		err := service.StartWorkspace(context.TODO(), createWorkspaceDTO.Id)

		require.Nil(t, err)
	})

	t.Run("StopWorkspace", func(t *testing.T) {
		provisioner.On("StopWorkspace", mock.Anything, &tg).Return(nil)

		err := service.StopWorkspace(context.TODO(), createWorkspaceDTO.Id)

		require.Nil(t, err)
	})

	t.Run("RemoveWorkspace", func(t *testing.T) {
		provisioner.On("DestroyWorkspace", mock.Anything, &tg).Return(nil)
		apiKeyService.On("Revoke", mock.Anything).Return(nil)

		err := service.RemoveWorkspace(context.TODO(), createWorkspaceDTO.Id)

		require.Nil(t, err)

		_, err = service.GetWorkspace(context.TODO(), createWorkspaceDTO.Id, true)
		require.Equal(t, workspaces.ErrWorkspaceNotFound, err)
	})

	t.Run("ForceRemoveWorkspace", func(t *testing.T) {
		err := workspaceStore.Save(ws)
		require.Nil(t, err)

		provisioner.On("DestroyWorkspace", mock.Anything, &tg).Return(nil)
		apiKeyService.On("Revoke", mock.Anything).Return(nil)

		err = service.ForceRemoveWorkspace(context.TODO(), createWorkspaceDTO.Id)

		require.Nil(t, err)

		_, err = service.GetWorkspace(context.TODO(), createWorkspaceDTO.Id, true)
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
		provisioner.AssertExpectations(t)
	})
}

func workspaceEquals(t *testing.T, req dto.CreateWorkspaceDTO, ws *workspace.Workspace, workspaceImage string) {
	t.Helper()

	// TODO: add more assertions
	require.Equal(t, req.Id, ws.Id)
	require.Equal(t, req.Name, ws.Name)
	require.Equal(t, req.TargetId, ws.TargetId)
	require.Equal(t, ws.Image, workspaceImage)
	require.Equal(t, ws.User, "daytona")
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
