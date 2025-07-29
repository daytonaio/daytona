// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"fmt"
	"strings"

	"github.com/daytonaio/daemon/pkg/gitprovider"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/capability"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

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

	if repo.Branch != "" {
		cloneOptions.ReferenceName = plumbing.ReferenceName("refs/heads/" + repo.Branch)
	}

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
	cloneCmd := []string{"git", "clone", "--single-branch"}

	// Only add branch flag if a specific branch is provided
	if repo.Branch != "" {
		cloneCmd = append(cloneCmd, "--branch", fmt.Sprintf("\"%s\"", repo.Branch))
	}

	cloneUrl := repo.Url

	// Default to https protocol if not specified
	if !strings.Contains(cloneUrl, "://") {
		cloneUrl = fmt.Sprintf("https://%s", cloneUrl)
	}

	if auth != nil {
		cloneUrl = fmt.Sprintf("%s://%s:%s@%s", strings.Split(cloneUrl, "://")[0], auth.Username, auth.Password, strings.SplitN(cloneUrl, "://", 2)[1])
	}

	cloneCmd = append(cloneCmd, cloneUrl, s.ProjectDir)

	if repo.Target == gitprovider.CloneTargetCommit {
		cloneCmd = append(cloneCmd, "&&", "cd", s.ProjectDir)
		cloneCmd = append(cloneCmd, "&&", "git", "checkout", repo.Sha)
	}

	return cloneCmd
}
