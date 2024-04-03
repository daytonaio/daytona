// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"os"

	"github.com/daytonaio/daytona/pkg/agent/config"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func CloneRepository(config *config.Config, project *serverapiclient.Project, authToken *string) error {
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

	if shouldCloneBranch(project) {
		cloneOptions.ReferenceName = plumbing.ReferenceName("refs/heads/" + *project.Repository.Branch)
	}

	_, err := git.PlainClone(config.ProjectDir, false, cloneOptions)
	if err != nil {
		return err
	}

	if shouldCheckoutSha(project) {
		repo, err := git.PlainOpen(config.ProjectDir)
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

func shouldCloneBranch(project *serverapiclient.Project) bool {
	if project.Repository.Branch == nil {
		return false
	}

	if project.Repository.Sha == nil {
		return true
	}

	return *project.Repository.Branch == *project.Repository.Sha
}

func shouldCheckoutSha(project *serverapiclient.Project) bool {
	if project.Repository.Sha == nil {
		return false
	}

	if project.Repository.Branch == nil {
		return true
	}

	return *project.Repository.Branch == *project.Repository.Sha
}
