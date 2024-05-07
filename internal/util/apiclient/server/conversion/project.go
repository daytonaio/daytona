// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package conversion

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/workspace"
)

func ToProject(projectDTO *serverapiclient.Project) *workspace.Project {
	if projectDTO == nil {
		return nil
	}

	repository := &gitprovider.GitRepository{
		Id:     *projectDTO.Repository.Id,
		Name:   *projectDTO.Repository.Name,
		Branch: projectDTO.Repository.Branch,
		Owner:  *projectDTO.Repository.Owner,
		Path:   projectDTO.Repository.Path,
		Sha:    *projectDTO.Repository.Sha,
		Source: *projectDTO.Repository.Source,
		Url:    *projectDTO.Repository.Url,
	}

	var projectState *workspace.ProjectState
	if projectDTO.State != nil {
		uptime := *projectDTO.State.Uptime
		projectState = &workspace.ProjectState{
			UpdatedAt: *projectDTO.State.UpdatedAt,
			Uptime:    uint64(uptime),
			GitStatus: ToGitStatus(projectDTO.State.GitStatus),
		}
	}

	project := &workspace.Project{
		Name:              *projectDTO.Name,
		Target:            *projectDTO.Target,
		WorkspaceId:       *projectDTO.WorkspaceId,
		Image:             *projectDTO.Image,
		User:              *projectDTO.User,
		PostStartCommands: projectDTO.PostStartCommands,
		Repository:        repository,
		State:             projectState,
	}

	if projectDTO.Repository.PrNumber != nil {
		prNumber := uint32(*projectDTO.Repository.PrNumber)
		project.Repository.PrNumber = &prNumber
	}

	return project
}

func ToGitStatus(gitStatusDTO *serverapiclient.GitStatus) *workspace.GitStatus {
	if gitStatusDTO == nil {
		return nil
	}

	files := []*workspace.FileStatus{}
	for _, fileDTO := range gitStatusDTO.FileStatus {
		staging := workspace.Status(string(*fileDTO.Staging))
		worktree := workspace.Status(string(*fileDTO.Worktree))
		file := &workspace.FileStatus{
			Name:     *fileDTO.Name,
			Extra:    *fileDTO.Extra,
			Staging:  staging,
			Worktree: worktree,
		}
		files = append(files, file)
	}

	return &workspace.GitStatus{
		CurrentBranch: *gitStatusDTO.CurrentBranch,
		Files:         files,
	}
}

func ToGitStatusDTO(gitStatus *workspace.GitStatus) *serverapiclient.GitStatus {
	if gitStatus == nil {
		return nil
	}

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
