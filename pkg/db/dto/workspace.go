// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/workspace"
)

type WorkspaceDTO struct {
	Id       string       `gorm:"primaryKey"`
	Name     string       `json:"name" gorm:"unique"`
	Target   string       `json:"target"`
	ApiKey   string       `json:"apiKey"`
	Projects []ProjectDTO `gorm:"serializer:json"`
}

func (w WorkspaceDTO) GetProject(name string) (*ProjectDTO, error) {
	for _, project := range w.Projects {
		if project.Name == name {
			return &project, nil
		}
	}

	return nil, errors.New("project not found")
}

func ToWorkspaceDTO(workspace *workspace.Workspace) WorkspaceDTO {
	workspaceDTO := WorkspaceDTO{
		Id:     workspace.Id,
		Name:   workspace.Name,
		Target: workspace.Target,
		ApiKey: workspace.ApiKey,
	}

	for _, project := range workspace.Projects {
		workspaceDTO.Projects = append(workspaceDTO.Projects, ToProjectDTO(project))
	}

	return workspaceDTO
}

func ToWorkspace(workspaceDTO WorkspaceDTO) *workspace.Workspace {
	workspace := workspace.Workspace{
		Id:     workspaceDTO.Id,
		Name:   workspaceDTO.Name,
		Target: workspaceDTO.Target,
		ApiKey: workspaceDTO.ApiKey,
	}

	for _, projectDTO := range workspaceDTO.Projects {
		workspace.Projects = append(workspace.Projects, ToProject(projectDTO))
	}

	return &workspace
}
