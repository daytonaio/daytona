// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/target/workspace"
	"github.com/daytonaio/daytona/pkg/target/workspace/buildconfig"
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

type WorkspaceStateDTO struct {
	UpdatedAt string        `json:"updatedAt"`
	Uptime    uint64        `json:"uptime"`
	GitStatus *GitStatusDTO `json:"gitStatus"`
}

type WorkspaceBuildDevcontainerDTO struct {
	FilePath string `json:"filePath"`
}

type WorkspaceBuildDTO struct {
	Devcontainer *WorkspaceBuildDevcontainerDTO `json:"devcontainer"`
}

type WorkspaceDTO struct {
	Name                string             `json:"name"`
	Image               string             `json:"image"`
	User                string             `json:"user"`
	Build               *WorkspaceBuildDTO `json:"build,omitempty" gorm:"serializer:json"`
	Repository          RepositoryDTO      `json:"repository" gorm:"serializer:json"`
	TargetId            string             `json:"targetId"`
	TargetConfig        string             `json:"targetConfig"`
	ApiKey              string             `json:"apiKey"`
	State               *WorkspaceStateDTO `json:"state,omitempty" gorm:"serializer:json"`
	GitProviderConfigId *string            `json:"gitProviderConfigId,omitempty"`
}

func ToWorkspaceDTO(workspace *workspace.Workspace) WorkspaceDTO {
	return WorkspaceDTO{
		Name:                workspace.Name,
		Image:               workspace.Image,
		User:                workspace.User,
		Build:               ToWorkspaceBuildDTO(workspace.BuildConfig),
		Repository:          ToRepositoryDTO(workspace.Repository),
		TargetId:            workspace.TargetId,
		TargetConfig:        workspace.TargetConfig,
		State:               ToWorkspaceStateDTO(workspace.State),
		ApiKey:              workspace.ApiKey,
		GitProviderConfigId: workspace.GitProviderConfigId,
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

func ToFileStatusDTO(status *workspace.FileStatus) *FileStatusDTO {
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

func ToGitStatusDTO(status *workspace.GitStatus) *GitStatusDTO {
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

func ToWorkspaceStateDTO(state *workspace.WorkspaceState) *WorkspaceStateDTO {
	if state == nil {
		return nil
	}

	return &WorkspaceStateDTO{
		UpdatedAt: state.UpdatedAt,
		Uptime:    state.Uptime,
		GitStatus: ToGitStatusDTO(state.GitStatus),
	}
}

func ToWorkspaceBuildDTO(build *buildconfig.BuildConfig) *WorkspaceBuildDTO {
	if build == nil {
		return nil
	}

	if build.Devcontainer == nil {
		return &WorkspaceBuildDTO{}
	}

	return &WorkspaceBuildDTO{
		Devcontainer: &WorkspaceBuildDevcontainerDTO{
			FilePath: build.Devcontainer.FilePath,
		},
	}
}

func ToWorkspace(workspaceDTO WorkspaceDTO) *workspace.Workspace {
	return &workspace.Workspace{
		Name:                workspaceDTO.Name,
		Image:               workspaceDTO.Image,
		User:                workspaceDTO.User,
		BuildConfig:         ToWorkspaceBuild(workspaceDTO.Build),
		Repository:          ToRepository(workspaceDTO.Repository),
		TargetId:            workspaceDTO.TargetId,
		TargetConfig:        workspaceDTO.TargetConfig,
		State:               ToWorkspaceState(workspaceDTO.State),
		ApiKey:              workspaceDTO.ApiKey,
		GitProviderConfigId: workspaceDTO.GitProviderConfigId,
	}
}

func ToFileStatus(statusDTO *FileStatusDTO) *workspace.FileStatus {
	if statusDTO == nil {
		return nil
	}

	return &workspace.FileStatus{
		Name:     statusDTO.Name,
		Extra:    statusDTO.Extra,
		Staging:  workspace.Status(statusDTO.Staging),
		Worktree: workspace.Status(statusDTO.Worktree),
	}
}

func ToGitStatus(statusDTO *GitStatusDTO) *workspace.GitStatus {
	if statusDTO == nil {
		return nil
	}

	status := &workspace.GitStatus{
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

func ToWorkspaceState(stateDTO *WorkspaceStateDTO) *workspace.WorkspaceState {
	if stateDTO == nil {
		return nil
	}

	return &workspace.WorkspaceState{
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

func ToWorkspaceBuild(buildDTO *WorkspaceBuildDTO) *buildconfig.BuildConfig {
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
