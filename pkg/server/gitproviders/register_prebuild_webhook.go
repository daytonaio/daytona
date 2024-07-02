// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitproviders

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/gitprovider"
)

func (s *GitProviderService) RegisterPrebuildWebhook(gitProviderId string, repo *gitprovider.GitRepository) error {
	gitProvider, err := s.GetGitProvider(gitProviderId)
	if err != nil {
		return fmt.Errorf("failed to get git provider: %s", err.Error())
	}

	err = gitProvider.RegisterPrebuildWebhook(repo)
	if err != nil {
		return fmt.Errorf("failed to get branches: %s", err.Error())
	}

	return nil
}
