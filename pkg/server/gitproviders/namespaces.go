// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitproviders

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/gitprovider"
)

func (s *GitProviderService) GetNamespaces(gitProviderId string) ([]*gitprovider.GitNamespace, error) {
	gitProvider, err := s.GetGitProvider(gitProviderId)
	if err != nil {
		return nil, fmt.Errorf("failed to get git provider: %w", err)
	}

	response, err := gitProvider.GetNamespaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get namespaces: %w", err)
	}

	return response, nil
}
