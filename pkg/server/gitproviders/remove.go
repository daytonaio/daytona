// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitproviders

import "github.com/daytonaio/daytona/pkg/target/project/config"

func (s *GitProviderService) RemoveGitProvider(gitProviderId string) error {
	gitProvider, err := s.configStore.Find(gitProviderId)
	if err != nil {
		return err
	}

	// Check if project configs need to be updated
	projectConfigs, err := s.projectConfigStore.List(&config.ProjectConfigFilter{
		GitProviderConfigId: &gitProviderId,
	})

	if err != nil {
		return err
	}

	for _, projectConfig := range projectConfigs {
		projectConfig.GitProviderConfigId = nil
		err = s.projectConfigStore.Save(projectConfig)
		if err != nil {
			return err
		}
	}

	return s.configStore.Delete(gitProvider)
}
