// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func (s *Service) Init(bare bool, initialBranch string) error {
	opts := &git.PlainInitOptions{
		Bare: bare,
	}
	if initialBranch != "" {
		opts.DefaultBranch = plumbing.NewBranchReferenceName(initialBranch)
	}

	_, err := git.PlainInitWithOptions(s.WorkDir, opts)
	return err
}
