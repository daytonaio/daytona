// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces_test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	t_logger "github.com/daytonaio/daytona/internal/testing/logger"
	t_targets "github.com/daytonaio/daytona/internal/testing/provider/targets"
	t_containerregistries "github.com/daytonaio/daytona/internal/testing/server/containerregistries"
	t_workspaces "github.com/daytonaio/daytona/internal/testing/server/workspaces"
	"github.com/daytonaio/daytona/internal/testing/server/workspaces/mocks"
	"github.com/daytonaio/daytona/pkg/apikey"
	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/logger"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/server/workspaces"
	"github.com/daytonaio/daytona/pkg/server/workspaces/dto"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const serverApiUrl = "http://localhost:3000"
const serverUrl = "http://localhost:3001"
const defaultProjectImage = "daytonaio/workspace-project:latest"
const defaultProjectUser = "daytona"

var target = provider.ProviderTarget{
	Name: "test-target",
	ProviderInfo: provider.ProviderInfo{
		Name:    "test-provider",
		Version: "test",
	},
	Options: "test-options",
}

var cr = containerregistry.ContainerRegistry{
	Server:   "test-server",
	Username: "test-username",
	Password: "test-password",
}

var createWorkspaceRequest = dto.CreateWorkspaceRequest{
	Name:   "test",
	Id:     "test",
	Target: target.Name,
	Projects: []dto.CreateWorkspaceRequestProject{
		{
			Id:   "project1",
			Name: "project1",
			Source: dto.CreateWorkspaceRequestProjectSource{
				Repository: &gitprovider.GitRepository{
					Id:   "123",
					Url:  "https://github.com/daytonaio/daytona",
					Name: "daytona",
				},
			},
		},
	},
}

var workspaceInfo = workspace.WorkspaceInfo{
	Name:             createWorkspaceRequest.Name,
	ProviderMetadata: "provider-metadata-test",
	Projects: []*workspace.ProjectInfo{
		{
			Name:             createWorkspaceRequest.Projects[0].Name,
			Created:          "1 min ago",
			IsRunning:        true,
			ProviderMetadata: "provider-metadata-test",
			WorkspaceId:      createWorkspaceRequest.Id,
		},
	},
}

func TestWorkspaceService(t *testing.T) {
	workspaceStore := t_workspaces.NewInMemoryWorkspaceStore()

	crStore := t_containerregistries.NewInMemoryContainerRegistryStore()
	err := crStore.Save(&cr)
	require.Nil(t, err)

	targetStore := t_targets.NewInMemoryTargetStore()
	err = targetStore.Save(&target)
	require.Nil(t, err)

	apiKeyService := mocks.NewMockApiKeyService()
	provisioner := mocks.NewMockProvisioner()

	logsDir := t.TempDir()

	service := workspaces.NewWorkspaceService(workspaces.WorkspaceServiceConfig{
		WorkspaceStore:         workspaceStore,
		TargetStore:            targetStore,
		ServerApiUrl:           serverApiUrl,
		ServerUrl:              serverUrl,
		ContainerRegistryStore: crStore,
		DefaultProjectImage:    defaultProjectImage,
		DefaultProjectUser:     defaultProjectUser,
		ApiKeyService:          apiKeyService,
		Provisioner:            provisioner,
		NewWorkspaceLogger: func(workspaceId string) logger.Logger {
			workspaceLogFilePath := filepath.Join(logsDir, workspaceId+".log")
			workspaceLogFile, err := os.Create(workspaceLogFilePath)
			if err != nil {
				t.Fatalf("failed to create workspace log file: %v", err)
			}
			return t_logger.NewPipeLogger(workspaceLogFile)
		},
		NewProjectLogger: func(workspaceId, projectName string) logger.Logger {
			projectLogFilePath := filepath.Join(logsDir, fmt.Sprintf("%s-%s.log", projectName, workspaceId))
			projectLogFile, err := os.Create(projectLogFilePath)
			if err != nil {
				t.Fatalf("failed to create project log file: %v", err)
			}
			return t_logger.NewPipeLogger(projectLogFile)
		},
		NewWorkspaceLogReader: func(workspaceId string) (io.Reader, error) {
			workspaceLogFilePath := filepath.Join(logsDir, workspaceId+".log")
			workspaceLogFile, err := os.Open(workspaceLogFilePath)
			if err != nil {
				return nil, err
			}
			return t_logger.NewPipeLogReader(workspaceLogFile)
		},
	})

	t.Run("CreateWorkspace", func(t *testing.T) {
		var containerRegistry *containerregistry.ContainerRegistry

		provisioner.On("CreateWorkspace", mock.Anything, &target).Return(nil)
		provisioner.On("StartWorkspace", mock.Anything, &target).Return(nil)

		for _, project := range createWorkspaceRequest.Projects {
			apiKeyService.On("Generate", apikey.ApiKeyTypeProject, fmt.Sprintf("%s/%s", createWorkspaceRequest.Id, project.Name)).Return(project.Name, nil)
		}
		provisioner.On("CreateProject", mock.Anything, &target, containerRegistry).Return(nil)
		provisioner.On("StartProject", mock.Anything, &target).Return(nil)

		workspace, err := service.CreateWorkspace(createWorkspaceRequest)

		require.Nil(t, err)
		require.NotNil(t, workspace)

		workspaceEquals(t, createWorkspaceRequest, workspace, defaultProjectImage)
	})

	t.Run("CreateWorkspace fails when workspace already exists", func(t *testing.T) {
		_, err := service.CreateWorkspace(createWorkspaceRequest)
		require.NotNil(t, err)
		require.Equal(t, workspaces.ErrWorkspaceAlreadyExists, err)
	})

	t.Run("CreateWorkspace fails name validation", func(t *testing.T) {
		invalidWorkspaceRequest := createWorkspaceRequest
		invalidWorkspaceRequest.Name = "invalid name"

		_, err := service.CreateWorkspace(invalidWorkspaceRequest)
		require.NotNil(t, err)
		require.Equal(t, workspaces.ErrInvalidWorkspaceName, err)
	})

	t.Run("GetWorkspace", func(t *testing.T) {
		provisioner.On("GetWorkspaceInfo", mock.Anything, &target).Return(&workspaceInfo, nil)

		workspace, err := service.GetWorkspace(createWorkspaceRequest.Id)

		require.Nil(t, err)
		require.NotNil(t, workspace)

		workspaceDtoEquals(t, createWorkspaceRequest, *workspace, workspaceInfo, defaultProjectImage, true)
	})

	t.Run("GetWorkspace fails when workspace not found", func(t *testing.T) {
		_, err := service.GetWorkspace("invalid-id")
		require.NotNil(t, err)
		require.Equal(t, workspaces.ErrWorkspaceNotFound, err)
	})

	t.Run("ListWorkspaces", func(t *testing.T) {
		verbose := false
		provisioner.On("GetWorkspaceInfo", mock.Anything, &target).Return(&workspaceInfo, nil)

		workspaces, err := service.ListWorkspaces(verbose)

		require.Nil(t, err)
		require.Len(t, workspaces, 1)

		workspace := workspaces[0]

		workspaceDtoEquals(t, createWorkspaceRequest, workspace, workspaceInfo, defaultProjectImage, verbose)
	})

	t.Run("ListWorkspaces - verbose", func(t *testing.T) {
		verbose := true
		provisioner.On("GetWorkspaceInfo", mock.Anything, &target).Return(&workspaceInfo, nil)

		workspaces, err := service.ListWorkspaces(verbose)

		require.Nil(t, err)
		require.Len(t, workspaces, 1)

		workspace := workspaces[0]

		workspaceDtoEquals(t, createWorkspaceRequest, workspace, workspaceInfo, defaultProjectImage, verbose)
	})

	t.Run("StartWorkspace", func(t *testing.T) {
		provisioner.On("StartWorkspace", mock.Anything, &target).Return(nil)
		provisioner.On("StartProject", mock.Anything, &target).Return(nil)

		err := service.StartWorkspace(createWorkspaceRequest.Id)

		require.Nil(t, err)
	})

	t.Run("StartProject", func(t *testing.T) {
		provisioner.On("StartWorkspace", mock.Anything, &target).Return(nil)
		provisioner.On("StartProject", mock.Anything, &target).Return(nil)

		err := service.StartProject(createWorkspaceRequest.Id, createWorkspaceRequest.Projects[0].Name)

		require.Nil(t, err)
	})

	t.Run("StopWorkspace", func(t *testing.T) {
		provisioner.On("StopWorkspace", mock.Anything, &target).Return(nil)
		provisioner.On("StopProject", mock.Anything, &target).Return(nil)

		err := service.StopWorkspace(createWorkspaceRequest.Id)

		require.Nil(t, err)
	})

	t.Run("StopProject", func(t *testing.T) {
		provisioner.On("StopWorkspace", mock.Anything, &target).Return(nil)
		provisioner.On("StopProject", mock.Anything, &target).Return(nil)

		err := service.StopProject(createWorkspaceRequest.Id, createWorkspaceRequest.Projects[0].Name)

		require.Nil(t, err)
	})

	t.Run("RemoveWorkspace", func(t *testing.T) {
		provisioner.On("DestroyWorkspace", mock.Anything, &target).Return(nil)
		provisioner.On("DestroyProject", mock.Anything, &target).Return(nil)
		apiKeyService.On("Revoke", mock.Anything).Return(nil)

		err := service.RemoveWorkspace(createWorkspaceRequest.Id)

		require.Nil(t, err)

		_, err = service.GetWorkspace(createWorkspaceRequest.Id)
		require.Equal(t, workspaces.ErrWorkspaceNotFound, err)
	})

	t.Cleanup(func() {
		apiKeyService.AssertExpectations(t)
		provisioner.AssertExpectations(t)
	})
}

func workspaceEquals(t *testing.T, req dto.CreateWorkspaceRequest, workspace *workspace.Workspace, projectImage string) {
	t.Helper()

	require.Equal(t, req.Id, workspace.Id)
	require.Equal(t, req.Name, workspace.Name)
	require.Equal(t, req.Target, workspace.Target)

	for i, project := range workspace.Projects {
		require.Equal(t, req.Projects[i].Name, project.Name)
		require.Equal(t, req.Projects[i].Source.Repository.Id, project.Repository.Id)
		require.Equal(t, req.Projects[i].Source.Repository.Url, project.Repository.Url)
		require.Equal(t, req.Projects[i].Source.Repository.Name, project.Repository.Name)
		require.Equal(t, project.ApiKey, project.Name)
		require.Equal(t, project.Target, req.Target)
		require.Equal(t, project.Image, projectImage)
	}
}

func workspaceDtoEquals(t *testing.T, req dto.CreateWorkspaceRequest, workspace dto.WorkspaceDTO, workspaceInfo workspace.WorkspaceInfo, projectImage string, verbose bool) {
	t.Helper()

	require.Equal(t, req.Id, workspace.Id)
	require.Equal(t, req.Name, workspace.Name)
	require.Equal(t, req.Target, workspace.Target)

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
		require.Equal(t, project.Target, req.Target)
		require.Equal(t, project.Image, projectImage)

		if verbose {
			require.Equal(t, workspace.Info.Projects[i].Name, workspaceInfo.Projects[i].Name)
			require.Equal(t, workspace.Info.Projects[i].Created, workspaceInfo.Projects[i].Created)
			require.Equal(t, workspace.Info.Projects[i].IsRunning, workspaceInfo.Projects[i].IsRunning)
			require.Equal(t, workspace.Info.Projects[i].ProviderMetadata, workspaceInfo.Projects[i].ProviderMetadata)
		}
	}
}
