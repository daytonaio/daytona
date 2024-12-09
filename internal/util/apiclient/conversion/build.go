// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package conversion

import (
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
)

// TODO: review - missing properties
func ToBuildDto(apiclientBuildDto *apiclient.BuildDTO) *services.BuildDTO {
	if apiclientBuildDto == nil {
		return nil
	}

	result := &services.BuildDTO{
		Build: models.Build{
			Id:              apiclientBuildDto.Id,
			Image:           apiclientBuildDto.Image,
			User:            apiclientBuildDto.User,
			ContainerConfig: models.ContainerConfig(apiclientBuildDto.ContainerConfig),
			EnvVars:         apiclientBuildDto.EnvVars,
			PrebuildId:      apiclientBuildDto.PrebuildId,
		},
	}

	repository := &gitprovider.GitRepository{
		Id:     apiclientBuildDto.Repository.Id,
		Name:   apiclientBuildDto.Repository.Name,
		Branch: apiclientBuildDto.Repository.Branch,
		Owner:  apiclientBuildDto.Repository.Owner,
		Path:   apiclientBuildDto.Repository.Path,
		Sha:    apiclientBuildDto.Repository.Sha,
		Source: apiclientBuildDto.Repository.Source,
		Url:    apiclientBuildDto.Repository.Url,
	}

	var buildConfig *models.BuildConfig
	if apiclientBuildDto.BuildConfig != nil {
		buildConfig = &models.BuildConfig{}
		if apiclientBuildDto.BuildConfig.Devcontainer != nil {
			buildConfig.Devcontainer = &models.DevcontainerConfig{
				FilePath: apiclientBuildDto.BuildConfig.Devcontainer.FilePath,
			}
		}
	}

	result.Repository = repository
	result.BuildConfig = buildConfig

	return result
}
