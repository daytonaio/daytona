// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitproviders

func (s *GitProviderService) RemoveGitProvider(gitProviderId string) error {
	gitProvider, err := s.configStore.Find(gitProviderId)
	if err != nil {
		return err
	}

	return s.configStore.Delete(gitProvider)
}
