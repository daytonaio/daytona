// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package types

type Project struct {
	Name        string            `json:"name"`
	Repository  *GitRepository    `json:"repository"`
	WorkspaceId string            `json:"workspaceId"`
	ApiKey      string            `json:"-"`
	Target      string            `json:"target"`
	EnvVars     map[string]string `json:"-"`
} // @name Project
type Workspace struct {
	Id       string     `json:"id"`
	Name     string     `json:"name"`
	Projects []*Project `json:"projects"`
	Target   string     `json:"target"`
} // @name Workspace

type ProjectInfo struct {
	Name             string `json:"name"`
	Created          string `json:"created"`
	Started          string `json:"started"`
	Finished         string `json:"finished"`
	IsRunning        bool   `json:"isRunning"`
	ProviderMetadata string `json:"providerMetadata,omitempty"`
	WorkspaceId      string `json:"workspaceId"`
} // @name ProjectInfo

type WorkspaceInfo struct {
	Name             string         `json:"name"`
	Projects         []*ProjectInfo `json:"projects"`
	ProviderMetadata string         `json:"providerMetadata,omitempty"`
} // @name WorkspaceInfo
