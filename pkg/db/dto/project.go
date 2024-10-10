// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/workspace/project"
	"github.com/daytonaio/daytona/pkg/workspace/project/buildconfig"
)

type RepositoryDTO struct {
	Id       string                  `json:"id"`
	Url      string                  `json:"url"`
	Name     string                  `json:"name"`
	Owner    string                  `json:"owner"`
	Sha      string                  `json:"sha"`
	Source   string                  `json:"source"`
	Branch   string                  `json:"branch"`
	PrNumber *uint32                 `json:"prNumber,omitempty"`
	Path     *string                 `json:"path,omitempty"`
	Target   gitprovider.CloneTarget `json:"cloneTarget,omitempty"`
}

type FileStatusDTO struct {
	Name     string `json:"name"`
	Extra    string `json:"extra"`
	Staging  string `json:"staging"`
	Worktree string `json:"worktree"`
}

type GitStatusDTO struct {
	CurrentBranch   string           `json:"currentBranch"`
	Files           []*FileStatusDTO `json:"fileStatus"`
	BranchPublished bool             `json:"branchPublished,omitempty"`
	Ahead           int32            `json:"ahead,omitempty"`
	Behind          int32            `json:"behind,omitempty"`
}

type ProjectStateDTO struct {
	UpdatedAt string        `json:"updatedAt"`
	Uptime    uint64        `json:"uptime"`
	GitStatus *GitStatusDTO `json:"gitStatus"`
}

type ProjectBuildDevcontainerDTO struct {
	FilePath string `json:"filePath"`
}

type ProjectBuildDTO struct {
	Devcontainer *ProjectBuildDevcontainerDTO `json:"devcontainer"`
}

type ProjectDTO struct {
	Name                string           `json:"name"`
	Image               string           `json:"image"`
	User                string           `json:"user"`
	Build               *ProjectBuildDTO `json:"build,omitempty" gorm:"serializer:json"`
	Repository          RepositoryDTO    `json:"repository" gorm:"serializer:json"`
	WorkspaceId         string           `json:"workspaceId"`
	Target              string           `json:"target"`
	ApiKey              string           `json:"apiKey"`
	State               *ProjectStateDTO `json:"state,omitempty" gorm:"serializer:json"`
	GitProviderConfigId *string          `json:"gitProviderConfigId,omitempty"`
}

func ToProjectDTO(project *project.Project) ProjectDTO {
	return ProjectDTO{
		Name:                project.Name,
		Image:               project.Image,
		User:                project.User,
		Build:               ToProjectBuildDTO(project.BuildConfig),
		Repository:          ToRepositoryDTO(project.Repository),
		WorkspaceId:         project.WorkspaceId,
		Target:              project.Target,
		State:               ToProjectStateDTO(project.State),
		ApiKey:              project.ApiKey,
		GitProviderConfigId: project.GitProviderConfigId,
	}
}

func ToRepositoryDTO(repo *gitprovider.GitRepository) RepositoryDTO {
	repoDTO := RepositoryDTO{
		Url:      repo.Url,
		Name:     repo.Name,
		Id:       repo.Id,
		Owner:    repo.Owner,
		Sha:      repo.Sha,
		Source:   repo.Source,
		Branch:   repo.Branch,
		PrNumber: repo.PrNumber,
		Path:     repo.Path,
		Target:   repo.Target,
	}

	return repoDTO
}

func ToFileStatusDTO(status *project.FileStatus) *FileStatusDTO {
	if status == nil {
		return nil
	}

	return &FileStatusDTO{
		Name:     status.Name,
		Extra:    status.Extra,
		Staging:  string(status.Staging),
		Worktree: string(status.Worktree),
	}
}

func ToGitStatusDTO(status *project.GitStatus) *GitStatusDTO {
	if status == nil {
		return nil
	}

	statusDTO := &GitStatusDTO{
		CurrentBranch:   status.CurrentBranch,
		BranchPublished: status.BranchPublished,
		Ahead:           int32(status.Ahead),
		Behind:          int32(status.Behind),
	}

	for _, file := range status.Files {
		statusDTO.Files = append(statusDTO.Files, ToFileStatusDTO(file))
	}

	return statusDTO
}

func ToProjectStateDTO(state *project.ProjectState) *ProjectStateDTO {
	if state == nil {
		return nil
	}

	return &ProjectStateDTO{
		UpdatedAt: state.UpdatedAt,
		Uptime:    state.Uptime,
		GitStatus: ToGitStatusDTO(state.GitStatus),
	}
}

func ToProjectBuildDTO(build *buildconfig.BuildConfig) *ProjectBuildDTO {
	if build == nil {
		return nil
	}

	if build.Devcontainer == nil {
		return &ProjectBuildDTO{}
	}

	return &ProjectBuildDTO{
		Devcontainer: &ProjectBuildDevcontainerDTO{
			FilePath: build.Devcontainer.FilePath,
		},
	}
}

func ToProject(projectDTO ProjectDTO) *project.Project {
	return &project.Project{
		Name:                projectDTO.Name,
		Image:               projectDTO.Image,
		User:                projectDTO.User,
		BuildConfig:         ToProjectBuild(projectDTO.Build),
		Repository:          ToRepository(projectDTO.Repository),
		WorkspaceId:         projectDTO.WorkspaceId,
		Target:              projectDTO.Target,
		State:               ToProjectState(projectDTO.State),
		ApiKey:              projectDTO.ApiKey,
		GitProviderConfigId: projectDTO.GitProviderConfigId,
	}
}

func ToFileStatus(statusDTO *FileStatusDTO) *project.FileStatus {
	if statusDTO == nil {
		return nil
	}

	return &project.FileStatus{
		Name:     statusDTO.Name,
		Extra:    statusDTO.Extra,
		Staging:  project.Status(statusDTO.Staging),
		Worktree: project.Status(statusDTO.Worktree),
	}
}

func ToGitStatus(statusDTO *GitStatusDTO) *project.GitStatus {
	if statusDTO == nil {
		return nil
	}

	status := &project.GitStatus{
		CurrentBranch:   statusDTO.CurrentBranch,
		BranchPublished: statusDTO.BranchPublished,
		Ahead:           int(statusDTO.Ahead),
		Behind:          int(statusDTO.Behind),
	}

	for _, file := range statusDTO.Files {
		status.Files = append(status.Files, ToFileStatus(file))
	}

	return status
}

func ToProjectState(stateDTO *ProjectStateDTO) *project.ProjectState {
	if stateDTO == nil {
		return nil
	}

	return &project.ProjectState{
		UpdatedAt: stateDTO.UpdatedAt,
		Uptime:    stateDTO.Uptime,
		GitStatus: ToGitStatus(stateDTO.GitStatus),
	}
}

func ToRepository(repoDTO RepositoryDTO) *gitprovider.GitRepository {
	repo := gitprovider.GitRepository{
		Url:      repoDTO.Url,
		Id:       repoDTO.Id,
		Name:     repoDTO.Name,
		Owner:    repoDTO.Owner,
		Branch:   repoDTO.Branch,
		Sha:      repoDTO.Sha,
		PrNumber: repoDTO.PrNumber,
		Source:   repoDTO.Source,
		Path:     repoDTO.Path,
		Target:   gitprovider.CloneTarget(repoDTO.Target),
	}

	return &repo
}

func ToProjectBuild(buildDTO *ProjectBuildDTO) *buildconfig.BuildConfig {
	if buildDTO == nil {
		return nil
	}

	if buildDTO.Devcontainer == nil {
		return &buildconfig.BuildConfig{}
	}

	return &buildconfig.BuildConfig{
		Devcontainer: &buildconfig.DevcontainerConfig{
			FilePath: buildDTO.Devcontainer.FilePath,
		},
	}
}
