// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package project

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/daytonaio/daytona/internal"
	"github.com/daytonaio/daytona/pkg/workspace/project/config"
)

type Project struct {
	config.ProjectConfig
	WorkspaceId string        `json:"workspaceId"`
	ApiKey      string        `json:"-"`
	Target      string        `json:"target"`
	State       *ProjectState `json:"state,omitempty"`
} // @name Project

type ProjectInfo struct {
	Name             string `json:"name"`
	Created          string `json:"created"`
	IsRunning        bool   `json:"isRunning"`
	ProviderMetadata string `json:"providerMetadata,omitempty"`
	WorkspaceId      string `json:"workspaceId"`
} // @name ProjectInfo

type ProjectState struct {
	UpdatedAt string     `json:"updatedAt"`
	Uptime    uint64     `json:"uptime"`
	GitStatus *GitStatus `json:"gitStatus"`
} // @name ProjectState

type GitStatus struct {
	CurrentBranch string        `json:"currentBranch"`
	Files         []*FileStatus `json:"fileStatus"`
} // @name GitStatus

type FileStatus struct {
	Name     string `json:"name"`
	Extra    string `json:"extra"`
	Staging  Status `json:"staging"`
	Worktree Status `json:"worktree"`
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
	ApiUrl    string
	ServerUrl string
	ClientId  string
}

func GetProjectEnvVars(project *Project, params ProjectEnvVarParams, telemetryEnabled bool) map[string]string {
	envVars := map[string]string{
		"DAYTONA_WS_ID":                     project.WorkspaceId,
		"DAYTONA_WS_PROJECT_NAME":           project.Name,
		"DAYTONA_WS_PROJECT_REPOSITORY_URL": project.Repository.Url,
		"DAYTONA_SERVER_API_KEY":            project.ApiKey,
		"DAYTONA_SERVER_VERSION":            internal.Version,
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

func GetProjectHostname(workspaceId string, projectName string) string {
	// Replace special chars with hyphen to form valid hostname
	// String resulting in consecutive hyphens is also valid
	projectName = strings.ReplaceAll(projectName, "_", "-")
	projectName = strings.ReplaceAll(projectName, "*", "-")
	projectName = strings.ReplaceAll(projectName, ".", "-")

	hostname := fmt.Sprintf("%s-%s", workspaceId, projectName)

	if len(hostname) > 63 {
		return hostname[:63]
	}

	return hostname
}

// GetConfigHash returns a SHA-256 hash of the project's build configuration, repository URL, and environment variables.
func (p *Project) GetConfigHash() (string, error) {
	buildJson, err := json.Marshal(p.Build)
	if err != nil {
		return "", err
	}

	//	todo: atm env vars contain workspace env provided by the server
	//		  this causes each workspace to have a different hash
	// envVarsJson, err := json.Marshal(p.EnvVars)
	// if err != nil {
	// 	return "", err
	// }

	data := string(buildJson) + p.Repository.Sha /* + string(envVarsJson)*/
	hash := sha256.Sum256([]byte(data))
	hashStr := hex.EncodeToString(hash[:])

	return hashStr, nil
}
