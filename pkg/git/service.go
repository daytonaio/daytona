// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"io"
	"os"
	"path/filepath"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

var MapStatus map[git.StatusCode]models.Status = map[git.StatusCode]models.Status{
	git.Unmodified:         models.Unmodified,
	git.Untracked:          models.Untracked,
	git.Modified:           models.Modified,
	git.Added:              models.Added,
	git.Deleted:            models.Deleted,
	git.Renamed:            models.Renamed,
	git.Copied:             models.Copied,
	git.UpdatedButUnmerged: models.UpdatedButUnmerged,
}

type IGitService interface {
	CloneRepository(repo *gitprovider.GitRepository, auth *http.BasicAuth) error
	CloneRepositoryCmd(repo *gitprovider.GitRepository, auth *http.BasicAuth) []string
	RepositoryExists() (bool, error)
	SetGitConfig(userData *gitprovider.GitUser, providerConfig *models.GitProviderConfig) error
	GetGitStatus() (*models.GitStatus, error)
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
