// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import "github.com/go-git/go-git/v5"

func (s *Service) Add(files []string) error {
	repo, err := git.PlainOpen(s.WorkspaceDir)
	if err != nil {
		return err
	}

	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	for _, file := range files {
		_, err = w.Add(file)
		if err != nil {
			return err
		}
	}

	return nil
}
