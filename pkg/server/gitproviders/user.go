// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitproviders

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/server/config"
)

func GetGitUser(gitProviderId string) (gitprovider.GitUser, error) {
	c, err := config.GetConfig()
	if err != nil {
		return gitprovider.GitUser{}, fmt.Errorf("failed to get config: %s", err.Error())
	}

	gitProvider := gitprovider.GetGitProvider(gitProviderId, c.GitProviders)
	if gitProvider == nil {
		return gitprovider.GitUser{}, fmt.Errorf("git provider not found")
	}

	user, err := gitProvider.GetUser()
	if err != nil {
		return gitprovider.GitUser{}, fmt.Errorf("failed to get user: %s", err.Error())
	}

	return user, nil
}

func GetGitUsernameFromToken(gitProviderData gitprovider.GitProvider) (string, error) {
	username, err := gitprovider.GetUsernameFromToken(gitProviderData.Id, gitProviderData.Token, gitProviderData.BaseApiUrl)
	if err != nil {
		return "", fmt.Errorf("failed to get username: %s", err.Error())
	}

	return username, nil
}
