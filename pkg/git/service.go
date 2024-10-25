// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"io"
	"os"
	"path/filepath"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/target/project"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

var MapStatus map[git.StatusCode]project.Status = map[git.StatusCode]project.Status{
	git.Unmodified:         project.Unmodified,
	git.Untracked:          project.Untracked,
	git.Modified:           project.Modified,
	git.Added:              project.Added,
	git.Deleted:            project.Deleted,
	git.Renamed:            project.Renamed,
	git.Copied:             project.Copied,
	git.UpdatedButUnmerged: project.UpdatedButUnmerged,
}

type IGitService interface {
	CloneRepository(repo *gitprovider.GitRepository, auth *http.BasicAuth) error
	CloneRepositoryCmd(repo *gitprovider.GitRepository, auth *http.BasicAuth) []string
	RepositoryExists() (bool, error)
	SetGitConfig(userData *gitprovider.GitUser, providerConfig *gitprovider.GitProviderConfig) error
	GetGitStatus() (*project.GitStatus, error)
}

type Service struct {
	ProjectDir        string
	GitConfigFileName string
	LogWriter         io.Writer
	OpenRepository    *git.Repository
}

func (s *Service) RepositoryExists() (bool, error) {
	_, err := os.Stat(filepath.Join(s.ProjectDir, ".git"))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
