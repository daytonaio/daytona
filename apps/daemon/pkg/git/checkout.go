// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func (s *Service) Checkout(branch string) error {
	r, err := git.PlainOpen(s.ProjectDir)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Try to checkout as a branch first
	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
	})

	if err != nil {
		// If branch checkout fails, try as a commit hash
		err = w.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(branch),
		})
		if err != nil {
			return fmt.Errorf("failed to checkout branch or commit '%s': %w", branch, err)
		}
	}

	return nil
}
