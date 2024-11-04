// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"io"
	"os"
	"path/filepath"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

var MapStatus map[git.StatusCode]workspace.Status = map[git.StatusCode]workspace.Status{
	git.Unmodified:         workspace.Unmodified,
	git.Untracked:          workspace.Untracked,
	git.Modified:           workspace.Modified,
	git.Added:              workspace.Added,
	git.Deleted:            workspace.Deleted,
	git.Renamed:            workspace.Renamed,
	git.Copied:             workspace.Copied,
	git.UpdatedButUnmerged: workspace.UpdatedButUnmerged,
}

type IGitService interface {
	CloneRepository(repo *gitprovider.GitRepository, auth *http.BasicAuth) error
	CloneRepositoryCmd(repo *gitprovider.GitRepository, auth *http.BasicAuth) []string
	RepositoryExists() (bool, error)
	SetGitConfig(userData *gitprovider.GitUser, providerConfig *gitprovider.GitProviderConfig) error
	GetGitStatus() (*workspace.GitStatus, error)
}

type Service struct {
	WorkspaceDir      string
	GitConfigFileName string
	LogWriter         io.Writer
	OpenRepository    *git.Repository
}

func (s *Service) RepositoryExists() (bool, error) {
	_, err := os.Stat(filepath.Join(s.WorkspaceDir, ".git"))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
