// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func (s *Service) Checkout(branch string) error {
	r, err := git.PlainOpen(s.WorkDir)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	branchErr := w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
	})
	if branchErr == nil {
		return nil
	}

	// Map missing-ref to the typed sentinel so the classifier emits GIT_BRANCH_NOT_FOUND.
	if errors.Is(branchErr, plumbing.ErrReferenceNotFound) {
		branchErr = fmt.Errorf("branch %q not found: %w", branch, git.ErrBranchNotFound)
	}

	// Only fall back to hash checkout when the input is plausible hex.
	if !looksLikeHash(branch) {
		return branchErr
	}

	if err := w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(branch),
	}); err != nil {
		if errors.Is(branchErr, git.ErrBranchNotFound) {
			return branchErr
		}
		return fmt.Errorf("failed to checkout branch or commit '%s': %w", branch, err)
	}

	return nil
}

func looksLikeHash(s string) bool {
	if len(s) < 4 || len(s) > 64 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}
