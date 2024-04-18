// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"encoding/base64"
	"errors"
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/daytonaio/daytona/internal"
	"github.com/daytonaio/daytona/pkg/gitprovider"
)

type ProjectImageUriParts struct {
	Server    string
	Username  *string
	ImageName string
}

type Project struct {
	Name        string                     `json:"name"`
	Image       string                     `json:"image"`
	Repository  *gitprovider.GitRepository `json:"repository"`
	WorkspaceId string                     `json:"workspaceId"`
	ApiKey      string                     `json:"-"`
	Target      string                     `json:"target"`
	EnvVars     map[string]string          `json:"-"`
	State       *ProjectState              `json:"state,omitempty"`
} // @name Project

func (p *Project) ParseImage() (*ProjectImageUriParts, error) {
	server := "docker.io"
	var user string
	image := ""

	parts := strings.Split(p.Image, "/")
	switch len(parts) {
	case 1:
		image = parts[0]
	case 2:
		user = parts[0]
		image = parts[1]
	case 3:
		server = parts[0]
		user = parts[1]
		image = parts[2]
	default:
		return nil, errors.New("invalid image uri")
	}

	return &ProjectImageUriParts{
		Server:    server,
		Username:  &user,
		ImageName: image,
	}, nil
}

type Workspace struct {
	Id       string     `json:"id"`
	Name     string     `json:"name"`
	Projects []*Project `json:"projects"`
	Target   string     `json:"target"`
} // @name Workspace

func (w *Workspace) GetProject(projectName string) (*Project, error) {
	for _, project := range w.Projects {
		if project.Name == projectName {
			return project, nil
		}
	}
	return nil, errors.New("project not found")
}

type ProjectState struct {
	UpdatedAt string `json:"updatedAt"`
	Uptime    uint64 `json:"uptime"`
} // @name ProjectState

type ProjectInfo struct {
	Name             string `json:"name"`
	Created          string `json:"created"`
	IsRunning        bool   `json:"isRunning"`
	ProviderMetadata string `json:"providerMetadata,omitempty"`
	WorkspaceId      string `json:"workspaceId"`
} // @name ProjectInfo

type WorkspaceInfo struct {
	Name             string         `json:"name"`
	Projects         []*ProjectInfo `json:"projects"`
	ProviderMetadata string         `json:"providerMetadata,omitempty"`
} // @name WorkspaceInfo

func GetProjectEnvVars(project *Project, apiUrl, serverUrl string) map[string]string {
	envVars := map[string]string{
		"DAYTONA_WS_ID":                     project.WorkspaceId,
		"DAYTONA_WS_PROJECT_NAME":           project.Name,
		"DAYTONA_WS_PROJECT_REPOSITORY_URL": project.Repository.Url,
		"DAYTONA_SERVER_API_KEY":            project.ApiKey,
		"DAYTONA_SERVER_VERSION":            internal.Version,
		"DAYTONA_SERVER_URL":                serverUrl,
		"DAYTONA_SERVER_API_URL":            apiUrl,
	}

	return envVars
}

func GetProjectHostname(workspaceId string, projectName string) string {
	h := fnv.New64()
	h.Write([]byte(fmt.Sprintf("%s-%s", workspaceId, projectName)))

	return base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprint(h.Sum64())))
}
