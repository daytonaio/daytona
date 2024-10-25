// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package project

import (
	"fmt"
	"strings"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/target/project/buildconfig"
)

type Project struct {
	Name                string                     `json:"name" validate:"required"`
	Image               string                     `json:"image" validate:"required"`
	User                string                     `json:"user" validate:"required"`
	BuildConfig         *buildconfig.BuildConfig   `json:"buildConfig,omitempty" validate:"optional"`
	Repository          *gitprovider.GitRepository `json:"repository" validate:"required"`
	EnvVars             map[string]string          `json:"envVars" validate:"required"`
	TargetId            string                     `json:"targetId" validate:"required"`
	ApiKey              string                     `json:"-"`
	TargetConfig        string                     `json:"targetConfig" validate:"required"`
	State               *ProjectState              `json:"state,omitempty" validate:"optional"`
	GitProviderConfigId *string                    `json:"gitProviderConfigId,omitempty" validate:"optional"`
} // @name Project

type ProjectInfo struct {
	Name             string `json:"name" validate:"required"`
	Created          string `json:"created" validate:"required"`
	IsRunning        bool   `json:"isRunning" validate:"required"`
	ProviderMetadata string `json:"providerMetadata,omitempty" validate:"optional"`
	TargetId         string `json:"targetId" validate:"required"`
} // @name ProjectInfo

type ProjectState struct {
	UpdatedAt string     `json:"updatedAt" validate:"required"`
	Uptime    uint64     `json:"uptime" validate:"required"`
	GitStatus *GitStatus `json:"gitStatus" validate:"optional"`
} // @name ProjectState

type GitStatus struct {
	CurrentBranch   string        `json:"currentBranch" validate:"required"`
	Files           []*FileStatus `json:"fileStatus" validate:"required"`
	BranchPublished bool          `json:"branchPublished" validate:"optional"`
	Ahead           int           `json:"ahead" validate:"optional"`
	Behind          int           `json:"behind" validate:"optional"`
} // @name GitStatus

type FileStatus struct {
	Name     string `json:"name" validate:"required"`
	Extra    string `json:"extra" validate:"required"`
	Staging  Status `json:"staging" validate:"required"`
	Worktree Status `json:"worktree" validate:"required"`
} // @name FileStatus

// Status status code of a file in the Worktree
type Status string // @name Status

const (
	Unmodified         Status = "Unmodified"
	Untracked          Status = "Untracked"
	Modified           Status = "Modified"
	Added              Status = "Added"
	Deleted            Status = "Deleted"
	Renamed            Status = "Renamed"
	Copied             Status = "Copied"
	UpdatedButUnmerged Status = "Updated but unmerged"
)

type ProjectEnvVarParams struct {
	ApiUrl        string
	ServerUrl     string
	ServerVersion string
	ClientId      string
}

func GetProjectEnvVars(project *Project, params ProjectEnvVarParams, telemetryEnabled bool) map[string]string {
	envVars := map[string]string{
		"DAYTONA_TARGET_ID":                 project.TargetId,
		"DAYTONA_PROJECT_NAME":              project.Name,
		"DAYTONA_WS_PROJECT_REPOSITORY_URL": project.Repository.Url,
		"DAYTONA_SERVER_API_KEY":            project.ApiKey,
		"DAYTONA_SERVER_VERSION":            params.ServerVersion,
		"DAYTONA_SERVER_URL":                params.ServerUrl,
		"DAYTONA_SERVER_API_URL":            params.ApiUrl,
		"DAYTONA_CLIENT_ID":                 params.ClientId,
		// (HOME) will be replaced at runtime
		"DAYTONA_AGENT_LOG_FILE_PATH": "(HOME)/.daytona-agent.log",
	}

	if telemetryEnabled {
		envVars["DAYTONA_TELEMETRY_ENABLED"] = "true"
	}

	return envVars
}

func GetProjectHostname(targetId string, projectName string) string {
	// Replace special chars with hyphen to form valid hostname
	// String resulting in consecutive hyphens is also valid
	projectName = strings.ReplaceAll(projectName, "_", "-")
	projectName = strings.ReplaceAll(projectName, "*", "-")
	projectName = strings.ReplaceAll(projectName, ".", "-")

	hostname := fmt.Sprintf("%s-%s", targetId, projectName)

	if len(hostname) > 63 {
		return hostname[:63]
	}

	return hostname
}
