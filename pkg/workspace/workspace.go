// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"errors"
)

type Workspace struct {
	Id       string     `json:"id"`
	Name     string     `json:"name"`
	Projects []*Project `json:"projects"`
	Target   string     `json:"target"`
} // @name Workspace

type WorkspaceInfo struct {
	Name             string         `json:"name"`
	Projects         []*ProjectInfo `json:"projects"`
	ProviderMetadata string         `json:"providerMetadata,omitempty"`
} // @name WorkspaceInfo

func (w *Workspace) GetProject(projectName string) (*Project, error) {
	for _, project := range w.Projects {
		if project.Name == projectName {
			return project, nil
		}
	}
	return nil, errors.New("project not found")
}
