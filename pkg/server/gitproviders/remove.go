// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitproviders

import "github.com/daytonaio/daytona/pkg/target/workspace/config"

func (s *GitProviderService) RemoveGitProvider(gitProviderId string) error {
	gitProvider, err := s.configStore.Find(gitProviderId)
	if err != nil {
		return err
	}

	// Check if workspace configs need to be updated
	workspaceConfigs, err := s.workspaceConfigStore.List(&config.WorkspaceConfigFilter{
		GitProviderConfigId: &gitProviderId,
	})

	if err != nil {
		return err
	}

	for _, workspaceConfig := range workspaceConfigs {
		workspaceConfig.GitProviderConfigId = nil
		err = s.workspaceConfigStore.Save(workspaceConfig)
		if err != nil {
			return err
		}
	}

	return s.configStore.Delete(gitProvider)
}
