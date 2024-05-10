// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"bytes"
	"io"
	"os"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"gopkg.in/ini.v1"
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
	CloneRepository(project *workspace.Project, auth *http.BasicAuth) error
	RepositoryExists(project *workspace.Project) (bool, error)
	SetGitConfig(userData *gitprovider.GitUser) error
	GetGitStatus() (*workspace.GitStatus, error)
}

type Service struct {
	ProjectDir        string
	GitConfigFileName string
	LogWriter         io.Writer
	OpenRepository    *git.Repository
}

func (s *Service) CloneRepository(project *workspace.Project, auth *http.BasicAuth) error {
	cloneOptions := &git.CloneOptions{
		URL:             project.Repository.Url,
		SingleBranch:    true,
		InsecureSkipTLS: true,
		Auth:            auth,
	}

	if s.LogWriter != nil {
		cloneOptions.Progress = s.LogWriter
	}

	if s.shouldCloneBranch(project) {
		cloneOptions.ReferenceName = plumbing.ReferenceName("refs/heads/" + *project.Repository.Branch)
	}

	_, err := git.PlainClone(s.ProjectDir, false, cloneOptions)
	if err != nil {
		return err
	}

	if s.shouldCheckoutSha(project) {
		repo, err := git.PlainOpen(s.ProjectDir)
		if err != nil {
			return err
		}

		w, err := repo.Worktree()
		if err != nil {
			return err
		}

		err = w.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(project.Repository.Sha),
		})
		if err != nil {
			return err
		}
	}

	return err
}

func (s *Service) RepositoryExists(project *workspace.Project) (bool, error) {
	_, err := os.Stat(s.ProjectDir)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *Service) SetGitConfig(userData *gitprovider.GitUser) error {
	gitConfigFileName := s.GitConfigFileName

	var gitConfigContent []byte
	gitConfigContent, err := os.ReadFile(gitConfigFileName)
	if err != nil {
		gitConfigContent = []byte{}
	}

	cfg, err := ini.Load(gitConfigContent)
	if err != nil {
		return err
	}

	if !cfg.HasSection("credential") {
		_, err := cfg.NewSection("credential")
		if err != nil {
			return err
		}
	}

	_, err = cfg.Section("credential").NewKey("helper", "/usr/local/bin/daytona git-cred")
	if err != nil {
		return err
	}

	if userData != nil {
		if !cfg.HasSection("user") {
			_, err := cfg.NewSection("user")
			if err != nil {
				return err
			}
		}

		_, err := cfg.Section("user").NewKey("name", userData.Name)
		if err != nil {
			return err
		}

		_, err = cfg.Section("user").NewKey("email", userData.Email)
		if err != nil {
			return err
		}
	}

	var buf bytes.Buffer
	_, err = cfg.WriteTo(&buf)
	if err != nil {
		return err
	}

	err = os.WriteFile(gitConfigFileName, buf.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetGitStatus() (*workspace.GitStatus, error) {
	repo, err := git.PlainOpen(s.ProjectDir)
	if err != nil {
		return nil, err
	}

	ref, err := repo.Head()
	if err != nil {
		return nil, err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	status, err := worktree.Status()
	if err != nil {
		return nil, err
	}

	files := []*workspace.FileStatus{}
	for path, file := range status {
		files = append(files, &workspace.FileStatus{
			Name:     path,
			Extra:    file.Extra,
			Staging:  MapStatus[file.Staging],
			Worktree: MapStatus[file.Worktree],
		})
	}

	return &workspace.GitStatus{
		CurrentBranch: ref.Name().Short(),
		Files:         files,
	}, nil
}

func (s *Service) shouldCloneBranch(project *workspace.Project) bool {
	if project.Repository.Branch == nil || *project.Repository.Branch == "" {
		return false
	}

	if project.Repository.Sha == "" {
		return true
	}

	return *project.Repository.Branch == project.Repository.Sha
}

func (s *Service) shouldCheckoutSha(project *workspace.Project) bool {
	if project.Repository.Sha == "" {
		return false
	}

	if project.Repository.Branch == nil || *project.Repository.Branch == "" {
		return true
	}

	return *project.Repository.Branch == project.Repository.Sha
}
