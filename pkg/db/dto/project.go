// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"strings"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/daytonaio/daytona/pkg/workspace/project"
	"github.com/daytonaio/daytona/pkg/workspace/project/build"
	"github.com/daytonaio/daytona/pkg/workspace/project/config"
)

type RepositoryDTO struct {
	Id       string  `json:"id"`
	Url      string  `json:"url"`
	Name     string  `json:"name"`
	Owner    string  `json:"owner"`
	Sha      string  `json:"sha"`
	Source   string  `json:"source"`
	Branch   *string `default:"main" json:"branch,omitempty"`
	PrNumber *uint32 `json:"prNumber,omitempty"`
	Path     *string `json:"path,omitempty"`
}

type FileStatusDTO struct {
	Name     string `json:"name"`
	Extra    string `json:"extra"`
	Staging  string `json:"staging"`
	Worktree string `json:"worktree"`
}

type GitStatusDTO struct {
	CurrentBranch string           `json:"currentBranch"`
	Files         []*FileStatusDTO `json:"fileStatus"`
}

type ProjectStateDTO struct {
	UpdatedAt string        `json:"updatedAt"`
	Uptime    uint64        `json:"uptime"`
	GitStatus *GitStatusDTO `json:"gitStatus"`
}

type ProjectBuildDevcontainerDTO struct {
	DevContainerFilePath string `json:"devContainerFilePath"`
}

type ProjectBuildDTO struct {
	Devcontainer *ProjectBuildDevcontainerDTO `json:"devcontainer"`
}

type ProjectDTO struct {
	Name        string           `json:"name"`
	Image       string           `json:"image"`
	User        string           `json:"user"`
	Build       *ProjectBuildDTO `json:"build,omitempty" gorm:"serializer:json"`
	Repository  RepositoryDTO    `json:"repository"`
	WorkspaceId string           `json:"workspaceId"`
	Target      string           `json:"target"`
	ApiKey      string           `json:"apiKey"`
	State       *ProjectStateDTO `json:"state,omitempty" gorm:"serializer:json"`
}

type ProjectConfigDTO struct {
	Name       string           `gorm:"primaryKey"`
	Image      string           `json:"image"`
	User       string           `json:"user"`
	Build      *ProjectBuildDTO `json:"build,omitempty" gorm:"serializer:json"`
	Repository RepositoryDTO    `gorm:"serializer:json"`
	EnvVars    string           `json:"envVars"`
	IsDefault  bool             `json:"isDefault"`
}

func ToProjectDTO(project *project.Project, workspace *workspace.Workspace) ProjectDTO {
	return ProjectDTO{
		Name:        project.Name,
		Image:       project.Image,
		User:        project.User,
		Build:       ToProjectBuildDTO(project.Build),
		Repository:  ToRepositoryDTO(project.Repository),
		WorkspaceId: project.WorkspaceId,
		Target:      project.Target,
		State:       ToProjectStateDTO(project.State),
		ApiKey:      workspace.ApiKey,
	}
}

func ToProjectConfigDTO(projectConfig *config.ProjectConfig) ProjectConfigDTO {
	return ProjectConfigDTO{
		Name:       projectConfig.Name,
		Image:      projectConfig.Image,
		User:       projectConfig.User,
		Build:      ToProjectBuildDTO(projectConfig.Build),
		Repository: ToRepositoryDTO(projectConfig.Repository),
		EnvVars:    ToEnvVarsString(projectConfig.EnvVars),
		IsDefault:  projectConfig.IsDefault,
	}
}

func ToProjectConfig(projectConfigDTO ProjectConfigDTO) *config.ProjectConfig {
	return &config.ProjectConfig{
		Name:       projectConfigDTO.Name,
		Image:      projectConfigDTO.Image,
		User:       projectConfigDTO.User,
		Build:      ToProjectBuild(projectConfigDTO.Build),
		Repository: ToRepository(projectConfigDTO.Repository),
		EnvVars:    ToEnvVarsMap(projectConfigDTO.EnvVars),
		IsDefault:  projectConfigDTO.IsDefault,
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
		CurrentBranch: status.CurrentBranch,
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

func ToProjectBuildDTO(build *build.ProjectBuildConfig) *ProjectBuildDTO {
	if build == nil {
		return nil
	}

	if build.Devcontainer == nil {
		return &ProjectBuildDTO{}
	}

	return &ProjectBuildDTO{
		Devcontainer: &ProjectBuildDevcontainerDTO{
			DevContainerFilePath: build.Devcontainer.FilePath,
		},
	}
}

func ToProject(projectDTO ProjectDTO) *project.Project {
	return &project.Project{
		ProjectConfig: config.ProjectConfig{
			Name:       projectDTO.Name,
			Image:      projectDTO.Image,
			User:       projectDTO.User,
			Build:      ToProjectBuild(projectDTO.Build),
			Repository: ToRepository(projectDTO.Repository),
		},
		WorkspaceId: projectDTO.WorkspaceId,
		Target:      projectDTO.Target,
		State:       ToProjectState(projectDTO.State),
		ApiKey:      projectDTO.ApiKey,
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
		CurrentBranch: statusDTO.CurrentBranch,
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
	}

	return &repo
}

func ToProjectBuild(buildDTO *ProjectBuildDTO) *build.ProjectBuildConfig {
	if buildDTO == nil {
		return nil
	}

	if buildDTO.Devcontainer == nil {
		return &build.ProjectBuildConfig{}
	}

	return &build.ProjectBuildConfig{
		Devcontainer: &build.DevcontainerConfig{
			FilePath: buildDTO.Devcontainer.DevContainerFilePath,
		},
	}
}

// TODO: handle special characters in env vars (e.g. newline, backslash, equals)

func ToEnvVarsString(envVars map[string]string) string {
	if envVars == nil {
		return ""
	}

	var builder strings.Builder
	for key, value := range envVars {
		builder.WriteString(key)
		builder.WriteString("=")
		builder.WriteString(value)
		builder.WriteString("\n")
	}

	result := builder.String()
	return strings.TrimSuffix(result, "\n")
}

func ToEnvVarsMap(envVars string) map[string]string {
	if envVars == "" {
		return nil
	}

	envMap := make(map[string]string)
	lines := strings.Split(envVars, "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	return envMap
}
