// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/workspace/project"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/capability"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"gopkg.in/ini.v1"
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
	SetGitConfig(userData *gitprovider.GitUser) error
	GetGitStatus() (*project.GitStatus, error)
}

type Service struct {
	ProjectDir        string
	GitConfigFileName string
	LogWriter         io.Writer
	OpenRepository    *git.Repository
}

func (s *Service) CloneRepository(repo *gitprovider.GitRepository, auth *http.BasicAuth) error {
	cloneOptions := &git.CloneOptions{
		URL:             repo.Url,
		SingleBranch:    true,
		InsecureSkipTLS: true,
		Auth:            auth,
	}

	if s.LogWriter != nil {
		cloneOptions.Progress = s.LogWriter
	}

	// Azure DevOps requires capabilities multi_ack / multi_ack_detailed,
	// which are not fully implemented and by default are included in
	// transport.UnsupportedCapabilities.
	//
	// This can be removed once go-git implements the git v2 protocol.
	transport.UnsupportedCapabilities = []capability.Capability{
		capability.ThinPack,
	}

	cloneOptions.ReferenceName = plumbing.ReferenceName("refs/heads/" + repo.Branch)

	_, err := git.PlainClone(s.ProjectDir, false, cloneOptions)
	if err != nil {
		return err
	}

	if repo.Target == gitprovider.CloneTargetCommit {
		r, err := git.PlainOpen(s.ProjectDir)
		if err != nil {
			return err
		}

		w, err := r.Worktree()
		if err != nil {
			return err
		}

		err = w.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(repo.Sha),
		})
		if err != nil {
			return err
		}
	}

	return err
}

func (s *Service) CloneRepositoryCmd(repo *gitprovider.GitRepository, auth *http.BasicAuth) []string {
	branch := fmt.Sprintf("\"%s\"", repo.Branch)
	cloneCmd := []string{"git", "clone", "--single-branch", "--branch", branch}

	if auth != nil {
		repoUrl := strings.TrimPrefix(repo.Url, "https://")
		repoUrl = strings.TrimPrefix(repoUrl, "http://")
		cloneCmd = append(cloneCmd, fmt.Sprintf("https://%s:%s@%s", auth.Username, auth.Password, repoUrl))
	} else {
		cloneCmd = append(cloneCmd, repo.Url)
	}

	cloneCmd = append(cloneCmd, s.ProjectDir)

	if repo.Target == gitprovider.CloneTargetCommit {
		cloneCmd = append(cloneCmd, "&&", "cd", s.ProjectDir)
		cloneCmd = append(cloneCmd, "&&", "git", "checkout", repo.Sha)
	}

	return cloneCmd
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

func (s *Service) GetGitStatus() (*project.GitStatus, error) {
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

	files := []*project.FileStatus{}
	for path, file := range status {
		files = append(files, &project.FileStatus{
			Name:     path,
			Extra:    file.Extra,
			Staging:  MapStatus[file.Staging],
			Worktree: MapStatus[file.Worktree],
		})
	}

	return &project.GitStatus{
		CurrentBranch: ref.Name().Short(),
		Files:         files,
	}, nil
}
