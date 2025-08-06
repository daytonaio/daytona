// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"fmt"

	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/go-git/go-git/v5"
)

func (s *Service) Push(auth *http.BasicAuth) error {
	repo, err := git.PlainOpen(s.ProjectDir)
	if err != nil {
		return err
	}

	ref, err := repo.Head()
	if err != nil {
		return err
	}

	options := &git.PushOptions{
		Auth: auth,
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("%s:%s", ref.Name(), ref.Name())),
		},
	}

	return repo.Push(options)
}
