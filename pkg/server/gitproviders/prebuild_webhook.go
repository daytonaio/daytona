// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitproviders

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/gitprovider"
)

func (s *GitProviderService) GetPrebuildWebhook(gitProviderId string, repo *gitprovider.GitRepository, endpointUrl string) (*string, error) {
	gitProvider, err := s.GetGitProvider(gitProviderId)
	if err != nil {
		return nil, fmt.Errorf("failed to get git provider: %s", err.Error())
	}

	id, err := gitProvider.GetPrebuildWebhook(repo, endpointUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook: %s", err.Error())
	}

	return id, nil
}

func (s *GitProviderService) RegisterPrebuildWebhook(gitProviderId string, repo *gitprovider.GitRepository, endpointUrl string) (string, error) {
	gitProvider, err := s.GetGitProvider(gitProviderId)
	if err != nil {
		return "", fmt.Errorf("failed to get git provider: %s", err.Error())
	}

	id, err := gitProvider.RegisterPrebuildWebhook(repo, endpointUrl)
	if err != nil {
		return "", fmt.Errorf("failed to register webhook: %s", err.Error())
	}

	return id, nil
}

func (s *GitProviderService) UnregisterPrebuildWebhook(gitProviderId string, repo *gitprovider.GitRepository, id string) error {
	gitProvider, err := s.GetGitProvider(gitProviderId)
	if err != nil {
		return fmt.Errorf("failed to get git provider: %s", err.Error())
	}

	err = gitProvider.UnregisterPrebuildWebhook(repo, id)
	if err != nil {
		return fmt.Errorf("failed to unregister webhook: %s", err.Error())
	}

	return nil
}
