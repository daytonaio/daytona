// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/workspace/project"
)

type Workspace struct {
	Id       string             `json:"id" validate:"required"`
	Name     string             `json:"name" validate:"required"`
	Projects []*project.Project `json:"projects" validate:"required"`
	Target   string             `json:"target" validate:"required"`
	ApiKey   string             `json:"-"`
	EnvVars  map[string]string  `json:"-"`
} // @name Workspace

type WorkspaceInfo struct {
	Name             string                 `json:"name" validate:"required"`
	Projects         []*project.ProjectInfo `json:"projects" validate:"required"`
	ProviderMetadata string                 `json:"providerMetadata,omitempty" validate:"optional"`
} // @name WorkspaceInfo

func (w *Workspace) GetProject(projectName string) (*project.Project, error) {
	for _, project := range w.Projects {
		if project.Name == projectName {
			return project, nil
		}
	}
	return nil, errors.New("project not found")
}

type WorkspaceEnvVarParams struct {
	ApiUrl        string
	ServerUrl     string
	ServerVersion string
	ClientId      string
}

func GetWorkspaceEnvVars(workspace *Workspace, params WorkspaceEnvVarParams, telemetryEnabled bool) map[string]string {
	envVars := map[string]string{
		"DAYTONA_WS_ID":          workspace.Id,
		"DAYTONA_SERVER_API_KEY": workspace.ApiKey,
		"DAYTONA_SERVER_VERSION": params.ServerVersion,
		"DAYTONA_SERVER_URL":     params.ServerUrl,
		"DAYTONA_SERVER_API_URL": params.ApiUrl,
		"DAYTONA_CLIENT_ID":      params.ClientId,
		// (HOME) will be replaced at runtime
		"DAYTONA_AGENT_LOG_FILE_PATH": "(HOME)/.daytona-agent.log",
	}

	if telemetryEnabled {
		envVars["DAYTONA_TELEMETRY_ENABLED"] = "true"
	}

	return envVars
}
