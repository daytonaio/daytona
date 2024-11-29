// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitproviders

import "context"

func (s *GitProviderService) RemoveGitProvider(gitProviderId string) error {
	ctx := context.Background()

	gitProvider, err := s.configStore.Find(gitProviderId)
	if err != nil {
		return err
	}

	err = s.detachWorkspaceTemplates(ctx, gitProvider.Id)
	if err != nil {
		return err
	}

	return s.configStore.Delete(gitProvider)
}
