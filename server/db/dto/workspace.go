package dto

import (
	"errors"

	"github.com/daytonaio/daytona/common/types"
)

type WorkspaceProvisionerDTO struct {
	Name    string `json:"name"`
	Profile string `json:"profile"`
}

type WorkspaceDTO struct {
	Id          string                  `gorm:"primaryKey"`
	Name        string                  `json:"name"`
	Provisioner WorkspaceProvisionerDTO `gorm:"serializer:json"`
	Projects    []ProjectDTO            `gorm:"serializer:json"`
}

type WorkspaceInfoDTO struct {
	Id          string                  `json:"id"`
	Name        string                  `json:"name"`
	Provisioner WorkspaceProvisionerDTO `json:"provisioner"`
	Projects    []ProjectInfoDTO        `json:"projects"`
	// TODO: rethink name
	ProvisionerMetadata interface{} `json:"provisionerMetadata"`
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
		Provisioner: WorkspaceProvisionerDTO{
			Name:    workspace.Provisioner.Name,
			Profile: workspace.Provisioner.Profile,
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
		Provisioner: &types.WorkspaceProvisioner{
			Name:    workspaceDTO.Provisioner.Name,
			Profile: workspaceDTO.Provisioner.Profile,
		},
	}

	for _, projectDTO := range workspaceDTO.Projects {
		workspace.Projects = append(workspace.Projects, ToProject(projectDTO))
	}

	return &workspace
}
