// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/daytonaio/daytona/internal"
	"github.com/daytonaio/daytona/pkg/gitprovider"
)

type ProjectBuildDevcontainer struct {
	DevContainerFilePath string `json:"devContainerFilePath"`
} // @name ProjectBuildDevcontainer

/*
type ProjectBuildDockerfile struct {
	Context    string            `json:"context"`
	Dockerfile string            `json:"dockerfile"`
	Args       map[string]string `json:"args"`
} // @name ProjectBuildDockerfile
*/

type ProjectBuild struct {
	Devcontainer *ProjectBuildDevcontainer `json:"devcontainer"`
	/*
		Dockerfile   *ProjectBuildDockerfile   `json:"dockerfile"`
	*/
} // @name ProjectBuild

type Project struct {
	Name               string                     `json:"name"`
	Image              string                     `json:"image"`
	User               string                     `json:"user"`
	Build              *ProjectBuild              `json:"build"`
	Repository         *gitprovider.GitRepository `json:"repository"`
	WorkspaceId        string                     `json:"workspaceId"`
	ApiKey             string                     `json:"-"`
	Target             string                     `json:"target"`
	EnvVars            map[string]string          `json:"-"`
	State              *ProjectState              `json:"state,omitempty"`
	PostCreateCommands []string                   `json:"postCreateCommands,omitempty"`
	PostStartCommands  []string                   `json:"postStartCommands,omitempty"`
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

func GetProjectEnvVars(project *Project, apiUrl, serverUrl string) map[string]string {
	envVars := map[string]string{
		"DAYTONA_WS_ID":                     project.WorkspaceId,
		"DAYTONA_WS_PROJECT_NAME":           project.Name,
		"DAYTONA_WS_PROJECT_REPOSITORY_URL": project.Repository.Url,
		"DAYTONA_SERVER_API_KEY":            project.ApiKey,
		"DAYTONA_SERVER_VERSION":            internal.Version,
		"DAYTONA_SERVER_URL":                serverUrl,
		"DAYTONA_SERVER_API_URL":            apiUrl,
		// (HOME) will be replaced at runtime
		"DAYTONA_AGENT_LOG_FILE_PATH": "(HOME)/.daytona-agent.log",
	}

	return envVars
}

func GetProjectHostname(workspaceId string, projectName string) string {
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
