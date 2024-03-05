// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import "github.com/daytonaio/daytona/pkg/types"

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
	ApiKey      string        `json:"apiKey"`
	Target      string        `json:"target"`
}

func ToProjectDTO(project *types.Project, workspace *types.Workspace) ProjectDTO {
	return ProjectDTO{
		Name:        project.Name,
		Repository:  ToRepositoryDTO(project.Repository),
		WorkspaceId: project.WorkspaceId,
		ApiKey:      project.ApiKey,
		Target:      project.Target,
	}
}

func ToRepositoryDTO(repo *types.Repository) RepositoryDTO {
	repoDTO := RepositoryDTO{
		Url: repo.Url,
	}

	if repo.Branch != "" {
		repoDTO.Branch = &repo.Branch
	}
	if repo.Sha != "" {
		repoDTO.SHA = &repo.Sha
	}
	if repo.Owner != "" {
		repoDTO.Owner = &repo.Owner
	}
	if repo.PrNumber != 0 {
		repoDTO.PrNumber = &repo.PrNumber
	}
	if repo.Source != "" {
		repoDTO.Source = &repo.Source
	}
	if repo.Path != "" {
		repoDTO.Path = &repo.Path
	}

	return repoDTO
}

func ToProject(projectDTO ProjectDTO) *types.Project {
	return &types.Project{
		Name:        projectDTO.Name,
		Repository:  ToRepository(projectDTO.Repository),
		WorkspaceId: projectDTO.WorkspaceId,
		ApiKey:      projectDTO.ApiKey,
		Target:      projectDTO.Target,
	}
}

func ToRepository(repoDTO RepositoryDTO) *types.Repository {
	repo := types.Repository{
		Url: repoDTO.Url,
	}

	if repoDTO.Branch != nil {
		repo.Branch = *repoDTO.Branch
	}
	if repoDTO.SHA != nil {
		repo.Sha = *repoDTO.SHA
	}
	if repoDTO.Owner != nil {
		repo.Owner = *repoDTO.Owner
	}
	if repoDTO.PrNumber != nil {
		repo.PrNumber = *repoDTO.PrNumber
	}
	if repoDTO.Source != nil {
		repo.Source = *repoDTO.Source
	}

	return &repo
}
