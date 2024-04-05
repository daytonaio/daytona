// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitproviders

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/gitprovider"
)

func (s *GitProviderService) GetRepositories(gitProviderId, namespaceId string) ([]*gitprovider.GitRepository, error) {
	gitProvider, err := s.GetGitProvider(gitProviderId)
	if err != nil {
		return nil, fmt.Errorf("failed to get git provider: %s", err.Error())
	}

	response, err := gitProvider.GetRepositories(namespaceId)
	if err != nil {
		return nil, fmt.Errorf("failed to get repositories: %s", err.Error())
	}

	return response, nil
}
