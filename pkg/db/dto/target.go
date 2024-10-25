// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/target"
)

type TargetDTO struct {
	Id           string       `gorm:"primaryKey"`
	Name         string       `json:"name" gorm:"unique"`
	TargetConfig string       `json:"config"`
	ApiKey       string       `json:"apiKey"`
	Projects     []ProjectDTO `gorm:"serializer:json"`
}

func (w TargetDTO) GetProject(name string) (*ProjectDTO, error) {
	for _, project := range w.Projects {
		if project.Name == name {
			return &project, nil
		}
	}

	return nil, errors.New("project not found")
}

func ToTargetDTO(target *target.Target) TargetDTO {
	targetDTO := TargetDTO{
		Id:           target.Id,
		Name:         target.Name,
		TargetConfig: target.TargetConfig,
		ApiKey:       target.ApiKey,
	}

	for _, project := range target.Projects {
		targetDTO.Projects = append(targetDTO.Projects, ToProjectDTO(project))
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

	for _, projectDTO := range targetDTO.Projects {
		target.Projects = append(target.Projects, ToProject(projectDTO))
	}

	return &target
}
