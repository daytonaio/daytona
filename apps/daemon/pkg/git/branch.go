// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func (s *Service) CreateBranch(name string) error {
	repo, err := git.PlainOpen(s.WorkDir)
	if err != nil {
		return err
	}

	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	return w.Checkout(&git.CheckoutOptions{
		Create: true,
		Branch: plumbing.NewBranchReferenceName(name),
	})
}

func (s *Service) ListBranches() ([]string, string, error) {
	repo, err := git.PlainOpen(s.WorkDir)
	if err != nil {
		return []string{}, "", err
	}

	branches, err := repo.Branches()
	if err != nil {
		return []string{}, "", err
	}

	branchList := []string{}
	err = branches.ForEach(func(ref *plumbing.Reference) error {
		branchList = append(branchList, ref.Name().Short())
		return nil
	})
	if err != nil {
		return branchList, "", err
	}

	// Read HEAD's symbolic target directly so the current branch is reported even
	// before the first commit (an unborn branch has no resolvable Head()).
	current := ""
	if head, headErr := repo.Reference(plumbing.HEAD, false); headErr == nil {
		switch {
		case head.Type() == plumbing.SymbolicReference && head.Target().IsBranch():
			current = head.Target().Short()
		case head.Name().IsBranch():
			current = head.Name().Short()
		}
	}

	return branchList, current, nil
}

func (s *Service) DeleteBranch(name string) error {
	repo, err := git.PlainOpen(s.WorkDir)
	if err != nil {
		return err
	}
	return repo.Storer.RemoveReference(plumbing.NewBranchReferenceName(name))
}
