// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"io"
	"os"
	"path/filepath"

	"github.com/daytonaio/daemon/pkg/gitprovider"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

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

var MapStatus map[git.StatusCode]Status = map[git.StatusCode]Status{
	git.Unmodified:         Unmodified,
	git.Untracked:          Untracked,
	git.Modified:           Modified,
	git.Added:              Added,
	git.Deleted:            Deleted,
	git.Renamed:            Renamed,
	git.Copied:             Copied,
	git.UpdatedButUnmerged: UpdatedButUnmerged,
}

type IGitService interface {
	CloneRepository(repo *gitprovider.GitRepository, auth *http.BasicAuth, insecureSkipTLS bool) error
	CloneRepositoryCLI(repo *gitprovider.GitRepository, auth *http.BasicAuth, insecureSkipTLS bool) error
	RepositoryExists() (bool, error)
	SetGitConfig(userData *gitprovider.GitUser, providerConfig *gitprovider.GitProviderConfig) error
	GetGitStatus() (*GitStatus, error)
}

type Service struct {
	WorkDir           string
	GitConfigFileName string
	LogWriter         io.Writer
	OpenRepository    *git.Repository
}

// Set to "true" to opt into the git-CLI clone path (bounded memory, needs `git` in PATH).
// Deprecated: prefer experimentalUseGitCLIEnv which covers all network git operations.
const experimentalUseGitCloneCLIEnv = "DAYTONA_EXPERIMENTAL_USE_GIT_CLONE_CLI"

// Set to "true" to opt into git-CLI paths for all network operations (clone, push, pull).
// Bounded memory via native git's mmap-based pack handling.
const experimentalUseGitCLIEnv = "DAYTONA_EXPERIMENTAL_USE_GIT_CLI"

// isGitCLIModeEnabled reports whether the umbrella `DAYTONA_EXPERIMENTAL_USE_GIT_CLI`
// env var is set. Push and pull only check this flag; clone additionally honors
// the legacy `DAYTONA_EXPERIMENTAL_USE_GIT_CLONE_CLI` for backward compatibility.
func isGitCLIModeEnabled() bool {
	return os.Getenv(experimentalUseGitCLIEnv) == "true"
}

func (s *Service) RepositoryExists() (bool, error) {
	_, err := os.Stat(filepath.Join(s.WorkDir, ".git"))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
