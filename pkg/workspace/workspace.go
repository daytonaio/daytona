// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"errors"

	"github.com/daytonaio/daytona/internal"
	"github.com/daytonaio/daytona/pkg/workspace/project"
)

type Workspace struct {
	Id       string             `json:"id"`
	Name     string             `json:"name"`
	Projects []*project.Project `json:"projects"`
	Target   string             `json:"target"`
	ApiKey   string             `json:"-"`
	EnvVars  map[string]string  `json:"-"`
} // @name Workspace

type WorkspaceInfo struct {
	Name             string                 `json:"name"`
	Projects         []*project.ProjectInfo `json:"projects"`
	ProviderMetadata string                 `json:"providerMetadata,omitempty"`
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
	ApiUrl    string
	ServerUrl string
	ClientId  string
}

func GetWorkspaceEnvVars(workspace *Workspace, params WorkspaceEnvVarParams, telemetryEnabled bool) map[string]string {
	envVars := map[string]string{
		"DAYTONA_WS_ID":          workspace.Id,
		"DAYTONA_SERVER_API_KEY": workspace.ApiKey,
		"DAYTONA_SERVER_VERSION": internal.Version,
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
