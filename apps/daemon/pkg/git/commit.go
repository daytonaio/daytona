// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import "github.com/go-git/go-git/v5"

func (s *Service) Commit(message string, options *git.CommitOptions) (string, error) {
	repo, err := git.PlainOpen(s.ProjectDir)
	if err != nil {
		return "", err
	}

	w, err := repo.Worktree()
	if err != nil {
		return "", err
	}

	commit, err := w.Commit(message, options)
	if err != nil {
		return "", err
	}

	return commit.String(), nil
}
