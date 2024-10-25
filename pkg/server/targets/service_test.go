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
	"github.com/daytonaio/daytona/pkg/target/project"
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

var createTargetDTO = dto.CreateTargetDTO{
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

var targetInfo = target.TargetInfo{
	Name:             createTargetDTO.Name,
	ProviderMetadata: "provider-metadata-test",
	Projects: []*project.ProjectInfo{
		{
			Name:             createTargetDTO.Projects[0].Name,
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

	projectConfigService := mocks.NewMockProjectConfigService()

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
		ProjectConfigService:     projectConfigService,
		DefaultProjectImage:      defaultProjectImage,
		DefaultProjectUser:       defaultProjectUser,
		ApiKeyService:            apiKeyService,
		Provisioner:              provisioner,
		LoggerFactory:            logs.NewLoggerFactory(&tgLogsDir, &buildLogsDir),
		GitProviderService:       gitProviderService,
	})

	t.Run("CreateTarget", func(t *testing.T) {
		var containerRegistry *containerregistry.ContainerRegistry

		containerRegistryService.On("FindByImageName", defaultProjectImage).Return(containerRegistry, containerregistry.ErrContainerRegistryNotFound)

		provisioner.On("CreateTarget", mock.Anything, &targetConfig).Return(nil)
		provisioner.On("StartTarget", mock.Anything, &targetConfig).Return(nil)

		apiKeyService.On("Generate", apikey.ApiKeyTypeTarget, createTargetDTO.Id).Return(createTargetDTO.Id, nil)
		gitProviderService.On("GetLastCommitSha", createTargetDTO.Projects[0].Source.Repository).Return("123", nil)

		for _, project := range createTargetDTO.Projects {
			apiKeyService.On("Generate", apikey.ApiKeyTypeProject, fmt.Sprintf("%s/%s", createTargetDTO.Id, project.Name)).Return(project.Name, nil)
		}
		provisioner.On("CreateProject", mock.Anything, &targetConfig, containerRegistry, &gitProviderConfig).Return(nil)
		provisioner.On("StartProject", mock.Anything, &targetConfig).Return(nil)

		gitProviderService.On("GetConfig", "github").Return(&gitProviderConfig, nil)

		target, err := service.CreateTarget(context.TODO(), createTargetDTO)

		require.Nil(t, err)
		require.NotNil(t, target)

		targetEquals(t, createTargetDTO, target, defaultProjectImage)
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

		targetDtoEquals(t, createTargetDTO, *target, targetInfo, defaultProjectImage, true)
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

		targetDtoEquals(t, createTargetDTO, target, targetInfo, defaultProjectImage, verbose)
	})

	t.Run("ListTargets - verbose", func(t *testing.T) {
		verbose := true
		provisioner.On("GetTargetInfo", mock.Anything, mock.Anything, &targetConfig).Return(&targetInfo, nil)

		targets, err := service.ListTargets(context.TODO(), verbose)

		require.Nil(t, err)
		require.Len(t, targets, 1)

		target := targets[0]

		targetDtoEquals(t, createTargetDTO, target, targetInfo, defaultProjectImage, verbose)
	})

	t.Run("StartTarget", func(t *testing.T) {
		provisioner.On("StartTarget", mock.Anything, &targetConfig).Return(nil)
		provisioner.On("StartProject", mock.Anything, &targetConfig).Return(nil)

		err := service.StartTarget(context.TODO(), createTargetDTO.Id)

		require.Nil(t, err)
	})

	t.Run("StartProject", func(t *testing.T) {
		provisioner.On("StartTarget", mock.Anything, &targetConfig).Return(nil)
		provisioner.On("StartProject", mock.Anything, &targetConfig).Return(nil)

		err := service.StartProject(context.TODO(), createTargetDTO.Id, createTargetDTO.Projects[0].Name)

		require.Nil(t, err)
	})

	t.Run("StopTarget", func(t *testing.T) {
		provisioner.On("StopTarget", mock.Anything, &targetConfig).Return(nil)
		provisioner.On("StopProject", mock.Anything, &targetConfig).Return(nil)

		err := service.StopTarget(context.TODO(), createTargetDTO.Id)

		require.Nil(t, err)
	})

	t.Run("StopProject", func(t *testing.T) {
		provisioner.On("StopTarget", mock.Anything, &targetConfig).Return(nil)
		provisioner.On("StopProject", mock.Anything, &targetConfig).Return(nil)

		err := service.StopProject(context.TODO(), createTargetDTO.Id, createTargetDTO.Projects[0].Name)

		require.Nil(t, err)
	})

	t.Run("RemoveTarget", func(t *testing.T) {
		provisioner.On("DestroyTarget", mock.Anything, &targetConfig).Return(nil)
		provisioner.On("DestroyProject", mock.Anything, &targetConfig).Return(nil)
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
		gitProviderService.On("GetLastCommitSha", createTargetDTO.Projects[0].Source.Repository).Return("123", nil)

		gitProviderService.On("ListConfigsForUrl", "https://github.com/daytonaio/daytona").Return([]*gitprovider.GitProviderConfig{&gitProviderConfig}, nil)

		for _, project := range createTargetDTO.Projects {
			apiKeyService.On("Generate", apikey.ApiKeyTypeProject, fmt.Sprintf("%s/%s", createTargetDTO.Id, project.Name)).Return(project.Name, nil)
		}
		provisioner.On("CreateProject", mock.Anything, &targetConfig, containerRegistry, &gitProviderConfig).Return(nil)
		provisioner.On("StartProject", mock.Anything, &targetConfig).Return(nil)

		_, _ = service.CreateTarget(context.TODO(), createTargetDTO)

		provisioner.On("DestroyTarget", mock.Anything, &targetConfig).Return(nil)
		provisioner.On("DestroyProject", mock.Anything, &targetConfig).Return(nil)
		apiKeyService.On("Revoke", mock.Anything).Return(nil)

		err = service.ForceRemoveTarget(context.TODO(), createTargetDTO.Id)

		require.Nil(t, err)

		_, err = service.GetTarget(context.TODO(), createTargetDTO.Id, true)
		require.Equal(t, targets.ErrTargetNotFound, err)
	})

	t.Run("SetProjectState", func(t *testing.T) {
		tg, err := service.CreateTarget(context.TODO(), createTargetDTO)
		require.Nil(t, err)

		projectName := tg.Projects[0].Name
		updatedAt := time.Now().Format(time.RFC1123)
		res, err := service.SetProjectState(tg.Id, projectName, &project.ProjectState{
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

func targetEquals(t *testing.T, req dto.CreateTargetDTO, target *target.Target, projectImage string) {
	t.Helper()

	require.Equal(t, req.Id, target.Id)
	require.Equal(t, req.Name, target.Name)
	require.Equal(t, req.TargetConfig, target.TargetConfig)

	for i, project := range target.Projects {
		require.Equal(t, req.Projects[i].Name, project.Name)
		require.Equal(t, req.Projects[i].Source.Repository.Id, project.Repository.Id)
		require.Equal(t, req.Projects[i].Source.Repository.Url, project.Repository.Url)
		require.Equal(t, req.Projects[i].Source.Repository.Name, project.Repository.Name)
		require.Equal(t, project.ApiKey, project.Name)
		require.Equal(t, project.TargetConfig, req.TargetConfig)
		require.Equal(t, project.Image, projectImage)
	}
}

func targetDtoEquals(t *testing.T, req dto.CreateTargetDTO, target dto.TargetDTO, targetInfo target.TargetInfo, projectImage string, verbose bool) {
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

	for i, project := range target.Projects {
		require.Equal(t, req.Projects[i].Name, project.Name)
		require.Equal(t, req.Projects[i].Source.Repository.Id, project.Repository.Id)
		require.Equal(t, req.Projects[i].Source.Repository.Url, project.Repository.Url)
		require.Equal(t, req.Projects[i].Source.Repository.Name, project.Repository.Name)
		require.Equal(t, project.ApiKey, project.Name)
		require.Equal(t, project.TargetConfig, req.TargetConfig)
		require.Equal(t, project.Image, projectImage)

		if verbose {
			require.Equal(t, target.Info.Projects[i].Name, targetInfo.Projects[i].Name)
			require.Equal(t, target.Info.Projects[i].Created, targetInfo.Projects[i].Created)
			require.Equal(t, target.Info.Projects[i].IsRunning, targetInfo.Projects[i].IsRunning)
			require.Equal(t, target.Info.Projects[i].ProviderMetadata, targetInfo.Projects[i].ProviderMetadata)
		}
	}
}
