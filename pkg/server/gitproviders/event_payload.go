// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitproviders

import (
	"fmt"
)

func (s *GitProviderService) ProcessWebhookEvent(gitProviderId string, webhookRequestPayload interface{}) error {
	gitProvider, err := s.GetGitProvider(gitProviderId)
	if err != nil {
		return fmt.Errorf("failed to get git provider: %s", err.Error())
	}

	err = gitProvider.ProcessWebhookEvent(webhookRequestPayload)
	if err != nil {
		return fmt.Errorf("failed to get branches: %s", err.Error())
	}

	return nil
}
