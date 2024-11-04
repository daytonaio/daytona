// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package conversion

import (
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	wc_dto "github.com/daytonaio/daytona/pkg/server/workspaceconfig/dto"
	workspace_dto "github.com/daytonaio/daytona/pkg/server/workspaces/dto"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/daytonaio/daytona/pkg/workspace/buildconfig"
	"github.com/daytonaio/daytona/pkg/workspace/config"
)

func ToWorkspace(workspaceDTO *apiclient.WorkspaceDTO) *workspace.Workspace {
	if workspaceDTO == nil {
		return nil
	}

	repository := &gitprovider.GitRepository{
		Id:     workspaceDTO.Repository.Id,
		Name:   workspaceDTO.Repository.Name,
		Branch: workspaceDTO.Repository.Branch,
		Owner:  workspaceDTO.Repository.Owner,
		Path:   workspaceDTO.Repository.Path,
		Sha:    workspaceDTO.Repository.Sha,
		Source: workspaceDTO.Repository.Source,
		Url:    workspaceDTO.Repository.Url,
	}

	var workspaceState *workspace.WorkspaceState
	if workspaceDTO.State != nil {
		uptime := workspaceDTO.State.Uptime
		workspaceState = &workspace.WorkspaceState{
			UpdatedAt: workspaceDTO.State.UpdatedAt,
			Uptime:    uint64(uptime),
			GitStatus: ToGitStatus(workspaceDTO.State.GitStatus),
		}
	}

	var workspaceBuild *buildconfig.BuildConfig
	if workspaceDTO.BuildConfig != nil {
		workspaceBuild = &buildconfig.BuildConfig{}
		if workspaceDTO.BuildConfig.Devcontainer != nil {
			workspaceBuild.Devcontainer = &buildconfig.DevcontainerConfig{
				FilePath: workspaceDTO.BuildConfig.Devcontainer.FilePath,
			}
		}
	}

	workspace := &workspace.Workspace{
		Id:                  workspaceDTO.Id,
		Name:                workspaceDTO.Name,
		Image:               workspaceDTO.Image,
		User:                workspaceDTO.User,
		BuildConfig:         workspaceBuild,
		Repository:          repository,
		TargetId:            workspaceDTO.TargetId,
		State:               workspaceState,
		GitProviderConfigId: workspaceDTO.GitProviderConfigId,
	}

	if workspaceDTO.Repository.PrNumber != nil {
		prNumber := uint32(*workspaceDTO.Repository.PrNumber)
		workspace.Repository.PrNumber = &prNumber
	}

	return workspace
}

func ToGitStatus(gitStatusDTO *apiclient.GitStatus) *workspace.GitStatus {
	if gitStatusDTO == nil {
		return nil
	}

	files := []*workspace.FileStatus{}
	for _, fileDTO := range gitStatusDTO.FileStatus {
		staging := workspace.Status(string(fileDTO.Staging))
		worktree := workspace.Status(string(fileDTO.Worktree))
		file := &workspace.FileStatus{
			Name:     fileDTO.Name,
			Extra:    fileDTO.Extra,
			Staging:  staging,
			Worktree: worktree,
		}
		files = append(files, file)
	}

	var ahead, behind int
	if gitStatusDTO.Ahead != nil {
		ahead = int(*gitStatusDTO.Ahead)
	}
	if gitStatusDTO.Behind != nil {
		behind = int(*gitStatusDTO.Behind)
	}

	var branchPublished bool
	if gitStatusDTO.BranchPublished != nil {
		branchPublished = *gitStatusDTO.BranchPublished
	}

	return &workspace.GitStatus{
		CurrentBranch:   gitStatusDTO.CurrentBranch,
		Files:           files,
		BranchPublished: branchPublished,
		Ahead:           ahead,
		Behind:          behind,
	}
}

func ToGitStatusDTO(gitStatus *workspace.GitStatus) *apiclient.GitStatus {
	if gitStatus == nil {
		return nil
	}

	fileStatusDTO := []apiclient.FileStatus{}
	for _, file := range gitStatus.Files {
		staging := apiclient.Status(string(file.Staging))
		worktree := apiclient.Status(string(file.Worktree))
		fileDTO := apiclient.FileStatus{
			Name:     file.Name,
			Extra:    file.Extra,
			Staging:  staging,
			Worktree: worktree,
		}
		fileStatusDTO = append(fileStatusDTO, fileDTO)
	}

	var ahead, behind *int32
	if gitStatus.Ahead != 0 {
		value := int32(gitStatus.Ahead)
		ahead = &value
	}
	if gitStatus.Behind != 0 {
		value := int32(gitStatus.Behind)
		behind = &value
	}
	var branchPublished *bool
	if gitStatus.BranchPublished {
		value := true
		branchPublished = &value
	}

	return &apiclient.GitStatus{
		CurrentBranch:   gitStatus.CurrentBranch,
		FileStatus:      fileStatusDTO,
		BranchPublished: branchPublished,
		Ahead:           ahead,
		Behind:          behind,
	}
}

func ToWorkspaceConfig(createWorkspaceConfigDto wc_dto.CreateWorkspaceConfigDTO) *config.WorkspaceConfig {
	result := &config.WorkspaceConfig{
		Name:                createWorkspaceConfigDto.Name,
		BuildConfig:         createWorkspaceConfigDto.BuildConfig,
		EnvVars:             createWorkspaceConfigDto.EnvVars,
		GitProviderConfigId: createWorkspaceConfigDto.GitProviderConfigId,
	}

	result.RepositoryUrl = createWorkspaceConfigDto.RepositoryUrl

	if createWorkspaceConfigDto.Image != nil {
		result.Image = *createWorkspaceConfigDto.Image
	}

	if createWorkspaceConfigDto.User != nil {
		result.User = *createWorkspaceConfigDto.User
	}

	return result
}

func CreateDtoToWorkspace(createWorkspaceDto workspace_dto.CreateWorkspaceDTO) *workspace.Workspace {
	w := &workspace.Workspace{
		Id:                  createWorkspaceDto.Id,
		Name:                createWorkspaceDto.Name,
		BuildConfig:         createWorkspaceDto.BuildConfig,
		Repository:          createWorkspaceDto.Source.Repository,
		EnvVars:             createWorkspaceDto.EnvVars,
		TargetId:            createWorkspaceDto.TargetId,
		GitProviderConfigId: createWorkspaceDto.GitProviderConfigId,
	}

	if createWorkspaceDto.Image != nil {
		w.Image = *createWorkspaceDto.Image
	}

	if createWorkspaceDto.User != nil {
		w.User = *createWorkspaceDto.User
	}

	return w
}

func CreateConfigDtoToWorkspace(createWorkspaceConfigDto wc_dto.CreateWorkspaceConfigDTO) *workspace.Workspace {
	return &workspace.Workspace{
		Name:                createWorkspaceConfigDto.Name,
		Image:               *createWorkspaceConfigDto.Image,
		User:                *createWorkspaceConfigDto.User,
		BuildConfig:         createWorkspaceConfigDto.BuildConfig,
		GitProviderConfigId: createWorkspaceConfigDto.GitProviderConfigId,
		Repository: &gitprovider.GitRepository{
			Url: createWorkspaceConfigDto.RepositoryUrl,
		},
		EnvVars: createWorkspaceConfigDto.EnvVars,
	}
}
