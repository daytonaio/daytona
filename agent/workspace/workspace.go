// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"errors"
	"regexp"
)

type WorkspaceProvisioner struct {
	Name    string `json:"name"`
	Profile string `json:"profile"`
}

type Workspace struct {
	Name        string               `gorm:"primaryKey"`
	Provisioner WorkspaceProvisioner `gorm:"serializer:json"`
	Projects    []Project            `gorm:"serializer:json"`
}

type CreateWorkspaceParams struct {
	Name string
	// Credentials  credentials.CredentialProvider
	Repositories []Repository
}

type WorkspaceInfo struct {
	Name        string               `json:"name"`
	Provisioner WorkspaceProvisioner `json:"provisioner"`
	Projects    []ProjectInfo        `json:"projects"`
	// TODO: rethink name
	ProvisionerMetadata interface{} `json:"provisionerMetadata"`
}

func New(params CreateWorkspaceParams) (*Workspace, error) {
	isAlphaNumeric := regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString
	if !isAlphaNumeric(params.Name) {
		return nil, errors.New("name is not a valid alphanumeric string")
	}

	w := Workspace{
		Name: params.Name,
	}
	w.Projects = []Project{}

	for _, repo := range params.Repositories {
		project := Project{
			Repository: repo,
			Workspace:  &w,
		}
		w.Projects = append(w.Projects, project)
	}

	return &w, nil
}

func (w Workspace) GetProject(name string) (*Project, error) {
	for _, project := range w.Projects {
		if project.Name == name {
			return &project, nil
		}
	}

	return nil, errors.New("project not found")
}
