// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package conversion

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/workspace"
)

func ToProjectDTO(project *workspace.Project) *serverapiclient.Project {
	repositoryDTO := &serverapiclient.GitRepository{
		Id:     &project.Repository.Id,
		Name:   &project.Repository.Name,
		Branch: project.Repository.Branch,
		Owner:  &project.Repository.Owner,
		Path:   project.Repository.Path,
		Sha:    &project.Repository.Sha,
		Source: &project.Repository.Source,
		Url:    &project.Repository.Url,
	}

	uptime := int32(project.State.Uptime)
	projectStateDTO := &serverapiclient.ProjectState{
		UpdatedAt: &project.State.UpdatedAt,
		Uptime:    &uptime,
		GitStatus: ToGitStatusDTO(project.State.GitStatus),
	}

	projectDto := &serverapiclient.Project{
		Name:              &project.Name,
		Target:            &project.Target,
		WorkspaceId:       &project.WorkspaceId,
		Image:             &project.Image,
		User:              &project.User,
		PostStartCommands: project.PostStartCommands,
		Repository:        repositoryDTO,
		State:             projectStateDTO,
	}

	if project.Repository.PrNumber != nil {
		prNumber := int32(*project.Repository.PrNumber)
		projectDto.Repository.PrNumber = &prNumber
	}

	return projectDto
}

func ToProject(project *serverapiclient.Project) *workspace.Project {
	repositoryDTO := &gitprovider.GitRepository{
		Id:     *project.Repository.Id,
		Name:   *project.Repository.Name,
		Branch: project.Repository.Branch,
		Owner:  *project.Repository.Owner,
		Path:   project.Repository.Path,
		Sha:    *project.Repository.Sha,
		Source: *project.Repository.Source,
		Url:    *project.Repository.Url,
	}

	uptime := *project.State.Uptime
	projectStateDTO := &workspace.ProjectState{
		UpdatedAt: *project.State.UpdatedAt,
		Uptime:    uint64(uptime),
		GitStatus: ToGitStatus(project.State.GitStatus),
	}

	projectDto := &workspace.Project{
		Name:              *project.Name,
		Target:            *project.Target,
		WorkspaceId:       *project.WorkspaceId,
		Image:             *project.Image,
		User:              *project.User,
		PostStartCommands: project.PostStartCommands,
		Repository:        repositoryDTO,
		State:             projectStateDTO,
	}

	if project.Repository.PrNumber != nil {
		prNumber := uint32(*project.Repository.PrNumber)
		projectDto.Repository.PrNumber = &prNumber
	}

	return projectDto
}

func ToGitStatusDTO(gitStatus *workspace.GitStatus) *serverapiclient.GitStatus {
	fileStatusDTO := []serverapiclient.FileStatus{}
	for _, file := range gitStatus.Files {
		staging := serverapiclient.Status(string(file.Staging))
		worktree := serverapiclient.Status(string(file.Worktree))
		fileDTO := serverapiclient.FileStatus{
			Name:     &file.Name,
			Extra:    &file.Extra,
			Staging:  &staging,
			Worktree: &worktree,
		}
		fileStatusDTO = append(fileStatusDTO, fileDTO)
	}

	return &serverapiclient.GitStatus{
		CurrentBranch: &gitStatus.CurrentBranch,
		FileStatus:    fileStatusDTO,
	}
}

func ToGitStatus(gitStatus *serverapiclient.GitStatus) *workspace.GitStatus {
	fileStatusDTO := []*workspace.FileStatus{}
	for _, file := range gitStatus.FileStatus {
		staging := workspace.Status(string(*file.Staging))
		worktree := workspace.Status(string(*file.Worktree))
		fileDTO := &workspace.FileStatus{
			Name:     *file.Name,
			Extra:    *file.Extra,
			Staging:  staging,
			Worktree: worktree,
		}
		fileStatusDTO = append(fileStatusDTO, fileDTO)
	}

	return &workspace.GitStatus{
		CurrentBranch: *gitStatus.CurrentBranch,
		Files:         fileStatusDTO,
	}
}
