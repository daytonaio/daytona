// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func (s *Service) CreateBranch(name string) error {
	repo, err := git.PlainOpen(s.ProjectDir)
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

func (s *Service) ListBranches() ([]string, error) {
	repo, err := git.PlainOpen(s.ProjectDir)
	if err != nil {
		return []string{}, err
	}

	branches, err := repo.Branches()
	if err != nil {
		return []string{}, err
	}

	var branchList []string
	err = branches.ForEach(func(ref *plumbing.Reference) error {
		branchList = append(branchList, ref.Name().Short())
		return nil
	})

	return branchList, err
}

func (s *Service) DeleteBranch(name string) error {
	repo, err := git.PlainOpen(s.ProjectDir)
	if err != nil {
		return err
	}
	return repo.Storer.RemoveReference(plumbing.NewBranchReferenceName(name))
}
