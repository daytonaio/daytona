// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitproviders

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/gitprovider"
)

func (s *GitProviderService) GetRepoPRs(gitProviderId, namespaceId, repositoryId string, options gitprovider.ListOptions) ([]*gitprovider.GitPullRequest, error) {
	gitProvider, err := s.GetGitProvider(gitProviderId)
	if err != nil {
		return nil, fmt.Errorf("failed to get git provider: %w", err)
	}

	response, err := gitProvider.GetRepoPRs(repositoryId, namespaceId, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get pull requests: %w", err)
	}

	return response, nil
}
