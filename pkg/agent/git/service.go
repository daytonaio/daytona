// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"bytes"
	"os"

	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"gopkg.in/ini.v1"
)

type Service struct {
	ProjectDir        string
	GitConfigFileName string
}

func (s *Service) CloneRepository(project *serverapiclient.Project, authToken *string) error {
	cloneOptions := &git.CloneOptions{
		URL:             *project.Repository.Url,
		Progress:        os.Stdout,
		SingleBranch:    true,
		InsecureSkipTLS: true,
	}

	if authToken != nil {
		cloneOptions.Auth = &http.BasicAuth{
			Username: "daytona",
			Password: *authToken,
		}
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
			Hash: plumbing.NewHash(*project.Repository.Sha),
		})
		if err != nil {
			return err
		}
	}

	return err
}

func (s *Service) RepositoryExists(project *serverapiclient.Project) (bool, error) {
	_, err := os.Stat(s.ProjectDir)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *Service) SetGitConfig(userData *serverapiclient.GitUser) error {
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

		if userData.Name != nil {
			_, err := cfg.Section("user").NewKey("name", *userData.Name)
			if err != nil {
				return err
			}
		}
		if userData.Email != nil {
			_, err := cfg.Section("user").NewKey("email", *userData.Email)
			if err != nil {
				return err
			}
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

func (s *Service) shouldCloneBranch(project *serverapiclient.Project) bool {
	if project.Repository.Branch == nil {
		return false
	}

	if project.Repository.Sha == nil {
		return true
	}

	return *project.Repository.Branch == *project.Repository.Sha
}

func (s *Service) shouldCheckoutSha(project *serverapiclient.Project) bool {
	if project.Repository.Sha == nil {
		return false
	}

	if project.Repository.Branch == nil {
		return true
	}

	return *project.Repository.Branch == *project.Repository.Sha
}
