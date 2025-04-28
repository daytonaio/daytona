// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/go-git/go-git/v5"
)

func (s *Service) Push(auth *http.BasicAuth) error {
	repo, err := git.PlainOpen(s.ProjectDir)
	if err != nil {
		return err
	}

	options := &git.PushOptions{
		Auth: auth,
	}

	return repo.Push(options)
}
