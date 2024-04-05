// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/workspace"
)

type RepositoryDTO struct {
	Url      string  `json:"url"`
	Branch   *string `default:"main" json:"branch,omitempty"`
	SHA      *string `json:"sha,omitempty"`
	Owner    *string `json:"owner,omitempty"`
	PrNumber *uint32 `json:"prNumber,omitempty"`
	Source   *string `json:"source,omitempty"`
	Path     *string `json:"path,omitempty"`
}

type ProjectDTO struct {
	Name        string        `json:"name"`
	Repository  RepositoryDTO `json:"repository"`
	WorkspaceId string        `json:"workspaceId"`
	Target      string        `json:"target"`
}

func ToProjectDTO(project *workspace.Project, workspace *workspace.Workspace) ProjectDTO {
	return ProjectDTO{
		Name:        project.Name,
		Repository:  ToRepositoryDTO(project.Repository),
		WorkspaceId: project.WorkspaceId,
		Target:      project.Target,
	}
}

func ToRepositoryDTO(repo *gitprovider.GitRepository) RepositoryDTO {
	repoDTO := RepositoryDTO{
		Url: repo.Url,
	}

	repoDTO.Branch = repo.Branch
	if repo.Sha != "" {
		repoDTO.SHA = &repo.Sha
	}
	if repo.Owner != "" {
		repoDTO.Owner = &repo.Owner
	}
	repoDTO.PrNumber = repo.PrNumber
	if repo.Source != "" {
		repoDTO.Source = &repo.Source
	}
	repoDTO.Path = repo.Path

	return repoDTO
}

func ToProject(projectDTO ProjectDTO) *workspace.Project {
	return &workspace.Project{
		Name:        projectDTO.Name,
		Repository:  ToRepository(projectDTO.Repository),
		WorkspaceId: projectDTO.WorkspaceId,
		Target:      projectDTO.Target,
	}
}

func ToRepository(repoDTO RepositoryDTO) *gitprovider.GitRepository {
	repo := gitprovider.GitRepository{
		Url: repoDTO.Url,
	}

	repo.Branch = repoDTO.Branch
	if repoDTO.SHA != nil {
		repo.Sha = *repoDTO.SHA
	}
	if repoDTO.Owner != nil {
		repo.Owner = *repoDTO.Owner
	}
	repo.PrNumber = repoDTO.PrNumber
	if repoDTO.Source != nil {
		repo.Source = *repoDTO.Source
	}

	return &repo
}
