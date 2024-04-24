// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package conversion

import (
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/workspace"
)

func ToProjectDTO(project *workspace.Project) *serverapiclient.Project {
	projectDto := &serverapiclient.Project{
		Name:              &project.Name,
		Target:            &project.Target,
		WorkspaceId:       &project.WorkspaceId,
		Image:             &project.Image,
		User:              &project.User,
		PostStartCommands: project.PostStartCommands,
		Repository: &serverapiclient.GitRepository{
			Id:     &project.Repository.Id,
			Name:   &project.Repository.Name,
			Branch: project.Repository.Branch,
			Owner:  &project.Repository.Owner,
			Path:   project.Repository.Path,
			Sha:    &project.Repository.Sha,
			Source: &project.Repository.Source,
			Url:    &project.Repository.Url,
		},
	}

	if project.Repository.PrNumber != nil {
		prNumber := int32(*project.Repository.PrNumber)
		projectDto.Repository.PrNumber = &prNumber
	}

	return projectDto
}
