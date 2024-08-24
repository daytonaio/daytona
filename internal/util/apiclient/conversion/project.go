// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package conversion

import (
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/server/projectconfig/dto"
	"github.com/daytonaio/daytona/pkg/workspace/project"
	"github.com/daytonaio/daytona/pkg/workspace/project/buildconfig"
	"github.com/daytonaio/daytona/pkg/workspace/project/config"
)

func ToProject(projectDTO *apiclient.Project) *project.Project {
	if projectDTO == nil {
		return nil
	}

	repository := &gitprovider.GitRepository{
		Id:     projectDTO.Repository.Id,
		Name:   projectDTO.Repository.Name,
		Branch: projectDTO.Repository.Branch,
		Owner:  projectDTO.Repository.Owner,
		Path:   projectDTO.Repository.Path,
		Sha:    projectDTO.Repository.Sha,
		Source: projectDTO.Repository.Source,
		Url:    projectDTO.Repository.Url,
	}

	var projectState *project.ProjectState
	if projectDTO.State != nil {
		uptime := projectDTO.State.Uptime
		projectState = &project.ProjectState{
			UpdatedAt: projectDTO.State.UpdatedAt,
			Uptime:    uint64(uptime),
			GitStatus: ToGitStatus(projectDTO.State.GitStatus),
		}
	}

	var projectBuild *buildconfig.ProjectBuildConfig
	if projectDTO.BuildConfig != nil {
		projectBuild = &buildconfig.ProjectBuildConfig{}
		if projectDTO.BuildConfig.Devcontainer != nil {
			projectBuild.Devcontainer = &buildconfig.DevcontainerConfig{
				FilePath: projectDTO.BuildConfig.Devcontainer.FilePath,
			}
		}
	}

	project := &project.Project{
		ProjectConfig: config.ProjectConfig{
			Name:        projectDTO.Name,
			Image:       projectDTO.Image,
			User:        projectDTO.User,
			BuildConfig: projectBuild,
			Repository:  repository,
		},
		Target:      projectDTO.Target,
		WorkspaceId: projectDTO.WorkspaceId,
		State:       projectState,
		Identity:    projectDTO.Identity,
	}

	if projectDTO.Repository.PrNumber != nil {
		prNumber := uint32(*projectDTO.Repository.PrNumber)
		project.Repository.PrNumber = &prNumber
	}

	return project
}

func ToGitStatus(gitStatusDTO apiclient.GitStatus) *project.GitStatus {
	files := []*project.FileStatus{}
	for _, fileDTO := range gitStatusDTO.FileStatus {
		staging := project.Status(string(fileDTO.Staging))
		worktree := project.Status(string(fileDTO.Worktree))
		file := &project.FileStatus{
			Name:     fileDTO.Name,
			Extra:    fileDTO.Extra,
			Staging:  staging,
			Worktree: worktree,
		}
		files = append(files, file)
	}

	return &project.GitStatus{
		CurrentBranch: gitStatusDTO.CurrentBranch,
		Files:         files,
	}
}

func ToGitStatusDTO(gitStatus *project.GitStatus) *apiclient.GitStatus {
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

	return &apiclient.GitStatus{
		CurrentBranch: gitStatus.CurrentBranch,
		FileStatus:    fileStatusDTO,
	}
}

func ToProjectConfig(createProjectConfigDto dto.CreateProjectConfigDTO) *config.ProjectConfig {
	result := &config.ProjectConfig{
		Name:        createProjectConfigDto.Name,
		BuildConfig: createProjectConfigDto.BuildConfig,
		EnvVars:     createProjectConfigDto.EnvVars,
	}

	if createProjectConfigDto.Source.Repository != nil {
		result.Repository = createProjectConfigDto.Source.Repository
	}

	if createProjectConfigDto.Image != nil {
		result.Image = *createProjectConfigDto.Image
	}

	if createProjectConfigDto.User != nil {
		result.User = *createProjectConfigDto.User
	}

	return result
}
