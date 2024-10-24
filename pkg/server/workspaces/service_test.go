// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	t_targetconfigs "github.com/daytonaio/daytona/internal/testing/provider/targetconfigs"
	t_workspaces "github.com/daytonaio/daytona/internal/testing/server/workspaces"
	"github.com/daytonaio/daytona/internal/testing/server/workspaces/mocks"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/apikey"
	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/server/workspaces"
	"github.com/daytonaio/daytona/pkg/server/workspaces/dto"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/daytonaio/daytona/pkg/workspace/project"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const serverApiUrl = "http://localhost:3986"
const serverUrl = "http://localhost:3987"
const defaultProjectUser = "daytona"
const defaultProjectImage = "daytonaio/workspace-project:latest"

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

var createWorkspaceDto = dto.CreateWorkspaceDTO{
	Name:         "test",
	Id:           "test",
	TargetConfig: targetConfig.Name,
	Projects: []dto.CreateProjectDTO{
		{
			Name:                "project1",
			GitProviderConfigId: &gitProviderConfig.Id,
			Source: dto.CreateProjectSourceDTO{
				Repository: &gitprovider.GitRepository{
					Id:     "123",
					Url:    "https://github.com/daytonaio/daytona",
					Name:   "daytona",
					Branch: "main",
					Sha:    "sha1",
				},
			},
			Image: util.Pointer(defaultProjectImage),
			User:  util.Pointer(defaultProjectUser),
		},
	},
}

var workspaceInfo = workspace.WorkspaceInfo{
	Name:             createWorkspaceDto.Name,
	ProviderMetadata: "provider-metadata-test",
	Projects: []*project.ProjectInfo{
		{
			Name:             createWorkspaceDto.Projects[0].Name,
			Created:          "1 min ago",
			IsRunning:        true,
			ProviderMetadata: "provider-metadata-test",
			WorkspaceId:      createWorkspaceDto.Id,
		},
	},
}

func TestWorkspaceService(t *testing.T) {
	workspaceStore := t_workspaces.NewInMemoryWorkspaceStore()

	containerRegistryService := mocks.NewMockContainerRegistryService()

	projectConfigService := mocks.NewMockProjectConfigService()

	targetConfigStore := t_targetconfigs.NewInMemoryTargetConfigStore()
	err := targetConfigStore.Save(&targetConfig)
	require.Nil(t, err)

	apiKeyService := mocks.NewMockApiKeyService()
	gitProviderService := mocks.NewMockGitProviderService()
	provisioner := mocks.NewMockProvisioner()

	wsLogsDir := t.TempDir()
	buildLogsDir := t.TempDir()

	service := workspaces.NewWorkspaceService(workspaces.WorkspaceServiceConfig{
		WorkspaceStore:           workspaceStore,
		TargetConfigStore:        targetConfigStore,
		ServerApiUrl:             serverApiUrl,
		ServerUrl:                serverUrl,
		ContainerRegistryService: containerRegistryService,
		ProjectConfigService:     projectConfigService,
		DefaultProjectImage:      defaultProjectImage,
		DefaultProjectUser:       defaultProjectUser,
		ApiKeyService:            apiKeyService,
		Provisioner:              provisioner,
		LoggerFactory:            logs.NewLoggerFactory(&wsLogsDir, &buildLogsDir),
		GitProviderService:       gitProviderService,
	})

	t.Run("CreateWorkspace", func(t *testing.T) {
		var containerRegistry *containerregistry.ContainerRegistry

		containerRegistryService.On("FindByImageName", defaultProjectImage).Return(containerRegistry, containerregistry.ErrContainerRegistryNotFound)

		provisioner.On("CreateWorkspace", mock.Anything, &targetConfig).Return(nil)
		provisioner.On("StartWorkspace", mock.Anything, &targetConfig).Return(nil)

		apiKeyService.On("Generate", apikey.ApiKeyTypeWorkspace, createWorkspaceDto.Id).Return(createWorkspaceDto.Id, nil)
		gitProviderService.On("GetLastCommitSha", createWorkspaceDto.Projects[0].Source.Repository).Return("123", nil)

		for _, project := range createWorkspaceDto.Projects {
			apiKeyService.On("Generate", apikey.ApiKeyTypeProject, fmt.Sprintf("%s/%s", createWorkspaceDto.Id, project.Name)).Return(project.Name, nil)
		}
		provisioner.On("CreateProject", mock.Anything, &targetConfig, containerRegistry, &gitProviderConfig).Return(nil)
		provisioner.On("StartProject", mock.Anything, &targetConfig).Return(nil)

		gitProviderService.On("GetConfig", "github").Return(&gitProviderConfig, nil)

		workspace, err := service.CreateWorkspace(context.TODO(), createWorkspaceDto)

		require.Nil(t, err)
		require.NotNil(t, workspace)

		workspaceEquals(t, createWorkspaceDto, workspace, defaultProjectImage)
	})

	t.Run("CreateWorkspace fails when workspace already exists", func(t *testing.T) {
		_, err := service.CreateWorkspace(context.TODO(), createWorkspaceDto)
		require.NotNil(t, err)
		require.Equal(t, workspaces.ErrWorkspaceAlreadyExists, err)
	})

	t.Run("CreateWorkspace fails name validation", func(t *testing.T) {
		invalidWorkspaceRequest := createWorkspaceDto
		invalidWorkspaceRequest.Name = "invalid name"

		_, err := service.CreateWorkspace(context.TODO(), invalidWorkspaceRequest)
		require.NotNil(t, err)
		require.Equal(t, workspaces.ErrInvalidWorkspaceName, err)
	})

	t.Run("GetWorkspace", func(t *testing.T) {
		provisioner.On("GetWorkspaceInfo", mock.Anything, mock.Anything, &targetConfig).Return(&workspaceInfo, nil)

		workspace, err := service.GetWorkspace(context.TODO(), createWorkspaceDto.Id, true)

		require.Nil(t, err)
		require.NotNil(t, workspace)

		workspaceDtoEquals(t, createWorkspaceDto, *workspace, workspaceInfo, defaultProjectImage, true)
	})

	t.Run("GetWorkspace fails when workspace not found", func(t *testing.T) {
		_, err := service.GetWorkspace(context.TODO(), "invalid-id", true)
		require.NotNil(t, err)
		require.Equal(t, workspaces.ErrWorkspaceNotFound, err)
	})

	t.Run("ListWorkspaces", func(t *testing.T) {
		verbose := false
		provisioner.On("GetWorkspaceInfo", mock.Anything, mock.Anything, &targetConfig).Return(&workspaceInfo, nil)

		workspaces, err := service.ListWorkspaces(context.TODO(), verbose)

		require.Nil(t, err)
		require.Len(t, workspaces, 1)

		workspace := workspaces[0]

		workspaceDtoEquals(t, createWorkspaceDto, workspace, workspaceInfo, defaultProjectImage, verbose)
	})

	t.Run("ListWorkspaces - verbose", func(t *testing.T) {
		verbose := true
		provisioner.On("GetWorkspaceInfo", mock.Anything, mock.Anything, &targetConfig).Return(&workspaceInfo, nil)

		workspaces, err := service.ListWorkspaces(context.TODO(), verbose)

		require.Nil(t, err)
		require.Len(t, workspaces, 1)

		workspace := workspaces[0]

		workspaceDtoEquals(t, createWorkspaceDto, workspace, workspaceInfo, defaultProjectImage, verbose)
	})

	t.Run("StartWorkspace", func(t *testing.T) {
		provisioner.On("StartWorkspace", mock.Anything, &targetConfig).Return(nil)
		provisioner.On("StartProject", mock.Anything, &targetConfig).Return(nil)

		err := service.StartWorkspace(context.TODO(), createWorkspaceDto.Id)

		require.Nil(t, err)
	})

	t.Run("StartProject", func(t *testing.T) {
		provisioner.On("StartWorkspace", mock.Anything, &targetConfig).Return(nil)
		provisioner.On("StartProject", mock.Anything, &targetConfig).Return(nil)

		err := service.StartProject(context.TODO(), createWorkspaceDto.Id, createWorkspaceDto.Projects[0].Name)

		require.Nil(t, err)
	})

	t.Run("StopWorkspace", func(t *testing.T) {
		provisioner.On("StopWorkspace", mock.Anything, &targetConfig).Return(nil)
		provisioner.On("StopProject", mock.Anything, &targetConfig).Return(nil)

		err := service.StopWorkspace(context.TODO(), createWorkspaceDto.Id)

		require.Nil(t, err)
	})

	t.Run("StopProject", func(t *testing.T) {
		provisioner.On("StopWorkspace", mock.Anything, &targetConfig).Return(nil)
		provisioner.On("StopProject", mock.Anything, &targetConfig).Return(nil)

		err := service.StopProject(context.TODO(), createWorkspaceDto.Id, createWorkspaceDto.Projects[0].Name)

		require.Nil(t, err)
	})

	t.Run("RemoveWorkspace", func(t *testing.T) {
		provisioner.On("DestroyWorkspace", mock.Anything, &targetConfig).Return(nil)
		provisioner.On("DestroyProject", mock.Anything, &targetConfig).Return(nil)
		apiKeyService.On("Revoke", mock.Anything).Return(nil)

		err := service.RemoveWorkspace(context.TODO(), createWorkspaceDto.Id)

		require.Nil(t, err)

		_, err = service.GetWorkspace(context.TODO(), createWorkspaceDto.Id, true)
		require.Equal(t, workspaces.ErrWorkspaceNotFound, err)
	})

	t.Run("ForceRemoveWorkspace", func(t *testing.T) {
		var containerRegistry *containerregistry.ContainerRegistry

		provisioner.On("CreateWorkspace", mock.Anything, &targetConfig).Return(nil)
		provisioner.On("StartWorkspace", mock.Anything, &targetConfig).Return(nil)

		apiKeyService.On("Generate", apikey.ApiKeyTypeWorkspace, createWorkspaceDto.Id).Return(createWorkspaceDto.Id, nil)
		gitProviderService.On("GetLastCommitSha", createWorkspaceDto.Projects[0].Source.Repository).Return("123", nil)

		gitProviderService.On("ListConfigsForUrl", "https://github.com/daytonaio/daytona").Return([]*gitprovider.GitProviderConfig{&gitProviderConfig}, nil)

		for _, project := range createWorkspaceDto.Projects {
			apiKeyService.On("Generate", apikey.ApiKeyTypeProject, fmt.Sprintf("%s/%s", createWorkspaceDto.Id, project.Name)).Return(project.Name, nil)
		}
		provisioner.On("CreateProject", mock.Anything, &targetConfig, containerRegistry, &gitProviderConfig).Return(nil)
		provisioner.On("StartProject", mock.Anything, &targetConfig).Return(nil)

		_, _ = service.CreateWorkspace(context.TODO(), createWorkspaceDto)

		provisioner.On("DestroyWorkspace", mock.Anything, &targetConfig).Return(nil)
		provisioner.On("DestroyProject", mock.Anything, &targetConfig).Return(nil)
		apiKeyService.On("Revoke", mock.Anything).Return(nil)

		err = service.ForceRemoveWorkspace(context.TODO(), createWorkspaceDto.Id)

		require.Nil(t, err)

		_, err = service.GetWorkspace(context.TODO(), createWorkspaceDto.Id, true)
		require.Equal(t, workspaces.ErrWorkspaceNotFound, err)
	})

	t.Run("SetProjectState", func(t *testing.T) {
		ws, err := service.CreateWorkspace(context.TODO(), createWorkspaceDto)
		require.Nil(t, err)

		projectName := ws.Projects[0].Name
		updatedAt := time.Now().Format(time.RFC1123)
		res, err := service.SetProjectState(ws.Id, projectName, &project.ProjectState{
			UpdatedAt: updatedAt,
			Uptime:    10,
			GitStatus: &project.GitStatus{
				CurrentBranch: "main",
			},
		})
		require.Nil(t, err)

		project, err := res.GetProject(projectName)
		require.Nil(t, err)
		require.Equal(t, "main", project.State.GitStatus.CurrentBranch)
	})

	t.Cleanup(func() {
		apiKeyService.AssertExpectations(t)
		provisioner.AssertExpectations(t)
	})
}

func workspaceEquals(t *testing.T, req dto.CreateWorkspaceDTO, workspace *workspace.Workspace, projectImage string) {
	t.Helper()

	require.Equal(t, req.Id, workspace.Id)
	require.Equal(t, req.Name, workspace.Name)
	require.Equal(t, req.TargetConfig, workspace.TargetConfig)

	for i, project := range workspace.Projects {
		require.Equal(t, req.Projects[i].Name, project.Name)
		require.Equal(t, req.Projects[i].Source.Repository.Id, project.Repository.Id)
		require.Equal(t, req.Projects[i].Source.Repository.Url, project.Repository.Url)
		require.Equal(t, req.Projects[i].Source.Repository.Name, project.Repository.Name)
		require.Equal(t, project.ApiKey, project.Name)
		require.Equal(t, project.TargetConfig, req.TargetConfig)
		require.Equal(t, project.Image, projectImage)
	}
}

func workspaceDtoEquals(t *testing.T, req dto.CreateWorkspaceDTO, workspace dto.WorkspaceDTO, workspaceInfo workspace.WorkspaceInfo, projectImage string, verbose bool) {
	t.Helper()

	require.Equal(t, req.Id, workspace.Id)
	require.Equal(t, req.Name, workspace.Name)
	require.Equal(t, req.TargetConfig, workspace.TargetConfig)

	if verbose {
		require.Equal(t, workspace.Info.Name, workspaceInfo.Name)
		require.Equal(t, workspace.Info.ProviderMetadata, workspaceInfo.ProviderMetadata)
	} else {
		require.Nil(t, workspace.Info)
	}

	for i, project := range workspace.Projects {
		require.Equal(t, req.Projects[i].Name, project.Name)
		require.Equal(t, req.Projects[i].Source.Repository.Id, project.Repository.Id)
		require.Equal(t, req.Projects[i].Source.Repository.Url, project.Repository.Url)
		require.Equal(t, req.Projects[i].Source.Repository.Name, project.Repository.Name)
		require.Equal(t, project.ApiKey, project.Name)
		require.Equal(t, project.TargetConfig, req.TargetConfig)
		require.Equal(t, project.Image, projectImage)

		if verbose {
			require.Equal(t, workspace.Info.Projects[i].Name, workspaceInfo.Projects[i].Name)
			require.Equal(t, workspace.Info.Projects[i].Created, workspaceInfo.Projects[i].Created)
			require.Equal(t, workspace.Info.Projects[i].IsRunning, workspaceInfo.Projects[i].IsRunning)
			require.Equal(t, workspace.Info.Projects[i].ProviderMetadata, workspaceInfo.Projects[i].ProviderMetadata)
		}
	}
}
