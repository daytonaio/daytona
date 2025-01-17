// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/go-git/go-git/v5"
)

func (s *Service) Log() ([]GitCommitInfo, error) {
	repo, err := git.PlainOpen(s.WorkspaceDir)
	if err != nil {
		return []GitCommitInfo{}, err
	}

	ref, err := repo.Head()
	if err != nil {
		return []GitCommitInfo{}, err
	}

	commits, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return []GitCommitInfo{}, err
	}

	var history []GitCommitInfo
	err = commits.ForEach(func(commit *object.Commit) error {
		history = append(history, GitCommitInfo{
			Hash:      commit.Hash.String(),
			Author:    commit.Author.Name,
			Email:     commit.Author.Email,
			Message:   commit.Message,
			Timestamp: commit.Author.When,
		})
		return nil
	})

	return history, err
}
