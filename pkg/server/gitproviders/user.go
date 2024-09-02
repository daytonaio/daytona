// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitproviders

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/gitprovider"
)

func (s *GitProviderService) GetGitUser(gitProviderId string) (*gitprovider.GitUser, error) {
	gitProvider, err := s.GetGitProvider(gitProviderId)
	if err != nil {
		return nil, fmt.Errorf("failed to get git provider: %w", err)
	}

	user, err := gitProvider.GetUser()
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}
