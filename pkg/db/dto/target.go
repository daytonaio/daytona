// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/target"
)

type TargetDTO struct {
	Id           string         `gorm:"primaryKey"`
	Name         string         `json:"name" gorm:"unique"`
	TargetConfig string         `json:"config"`
	ApiKey       string         `json:"apiKey"`
	Workspaces   []WorkspaceDTO `gorm:"serializer:json"`
}

func (w TargetDTO) GetWorkspace(name string) (*WorkspaceDTO, error) {
	for _, workspace := range w.Workspaces {
		if workspace.Name == name {
			return &workspace, nil
		}
	}

	return nil, errors.New("workspace not found")
}

func ToTargetDTO(target *target.Target) TargetDTO {
	targetDTO := TargetDTO{
		Id:           target.Id,
		Name:         target.Name,
		TargetConfig: target.TargetConfig,
		ApiKey:       target.ApiKey,
	}

	for _, workspace := range target.Workspaces {
		targetDTO.Workspaces = append(targetDTO.Workspaces, ToWorkspaceDTO(workspace))
	}

	return targetDTO
}

func ToTarget(targetDTO TargetDTO) *target.Target {
	target := target.Target{
		Id:           targetDTO.Id,
		Name:         targetDTO.Name,
		TargetConfig: targetDTO.TargetConfig,
		ApiKey:       targetDTO.ApiKey,
	}

	for _, workspaceDTO := range targetDTO.Workspaces {
		target.Workspaces = append(target.Workspaces, ToWorkspace(workspaceDTO))
	}

	return &target
}
