// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package conversion

import (
	"time"

	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/models"
	wc_dto "github.com/daytonaio/daytona/pkg/server/workspaceconfigs/dto"
)

func ToWorkspace(workspaceDTO *apiclient.WorkspaceDTO) *models.Workspace {
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

	var workspaceMetadata *models.WorkspaceMetadata
	if workspaceDTO.Metadata != nil {
		updatedAt, err := time.Parse(time.RFC3339, workspaceDTO.Metadata.UpdatedAt)
		if err != nil {
			updatedAt = time.Unix(0, 0)
		}
		uptime := workspaceDTO.Metadata.Uptime
		workspaceMetadata = &models.WorkspaceMetadata{
			UpdatedAt: updatedAt,
			Uptime:    uint64(uptime),
			GitStatus: ToGitStatus(workspaceDTO.Metadata.GitStatus),
		}
	}

	var workspaceBuild *models.BuildConfig
	if workspaceDTO.BuildConfig != nil {
		workspaceBuild = &models.BuildConfig{}
		if workspaceDTO.BuildConfig.Devcontainer != nil {
			workspaceBuild.Devcontainer = &models.DevcontainerConfig{
				FilePath: workspaceDTO.BuildConfig.Devcontainer.FilePath,
			}
		}
	}

	workspace := &models.Workspace{
		Id:                  workspaceDTO.Id,
		Name:                workspaceDTO.Name,
		Image:               workspaceDTO.Image,
		User:                workspaceDTO.User,
		BuildConfig:         workspaceBuild,
		Repository:          repository,
		TargetId:            workspaceDTO.TargetId,
		Metadata:            workspaceMetadata,
		GitProviderConfigId: workspaceDTO.GitProviderConfigId,
	}

	if workspaceDTO.Repository.PrNumber != nil {
		prNumber := uint32(*workspaceDTO.Repository.PrNumber)
		workspace.Repository.PrNumber = &prNumber
	}

	return workspace
}

func ToGitStatus(gitStatusDTO apiclient.GitStatus) *models.GitStatus {
	files := []*models.FileStatus{}
	for _, fileDTO := range gitStatusDTO.FileStatus {
		staging := models.Status(string(fileDTO.Staging))
		worktree := models.Status(string(fileDTO.Worktree))
		file := &models.FileStatus{
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

	return &models.GitStatus{
		CurrentBranch:   gitStatusDTO.CurrentBranch,
		Files:           files,
		BranchPublished: branchPublished,
		Ahead:           ahead,
		Behind:          behind,
	}
}

func ToGitStatusDTO(gitStatus *models.GitStatus) *apiclient.GitStatus {
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

func ToWorkspaceConfig(createWorkspaceConfigDto wc_dto.CreateWorkspaceConfigDTO) *models.WorkspaceConfig {
	result := &models.WorkspaceConfig{
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
