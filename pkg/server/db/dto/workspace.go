package dto

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/types"
)

type WorkspaceProviderDTO struct {
	Name    string `json:"name"`
	Profile string `json:"profile"`
}

type WorkspaceDTO struct {
	Id       string               `gorm:"primaryKey"`
	Name     string               `json:"name"`
	Provider WorkspaceProviderDTO `gorm:"serializer:json"`
	Projects []ProjectDTO         `gorm:"serializer:json"`
}

type WorkspaceInfoDTO struct {
	Id       string               `json:"id"`
	Name     string               `json:"name"`
	Provider WorkspaceProviderDTO `json:"provider"`
	Projects []ProjectInfoDTO     `json:"projects"`
	// TODO: rethink name
	ProviderMetadata interface{} `json:"providerMetadata"`
}

func (w WorkspaceDTO) GetProject(name string) (*ProjectDTO, error) {
	for _, project := range w.Projects {
		if project.Name == name {
			return &project, nil
		}
	}

	return nil, errors.New("project not found")
}

func ToWorkspaceDTO(workspace *types.Workspace) WorkspaceDTO {
	workspaceDTO := WorkspaceDTO{
		Id:   workspace.Id,
		Name: workspace.Name,
		Provider: WorkspaceProviderDTO{
			Name:    workspace.Provider.Name,
			Profile: workspace.Provider.Profile,
		},
	}

	for _, project := range workspace.Projects {
		workspaceDTO.Projects = append(workspaceDTO.Projects, ToProjectDTO(project, workspace))
	}

	return workspaceDTO
}

func ToWorkspace(workspaceDTO WorkspaceDTO) *types.Workspace {
	workspace := types.Workspace{
		Id:   workspaceDTO.Id,
		Name: workspaceDTO.Name,
		Provider: &types.WorkspaceProvider{
			Name:    workspaceDTO.Provider.Name,
			Profile: workspaceDTO.Provider.Profile,
		},
	}

	for _, projectDTO := range workspaceDTO.Projects {
		workspace.Projects = append(workspace.Projects, ToProject(projectDTO))
	}

	return &workspace
}
