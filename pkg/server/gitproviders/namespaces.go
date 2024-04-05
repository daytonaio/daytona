// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitproviders

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/server/config"
)

func GetNamespaces(gitProviderId string) ([]gitprovider.GitNamespace, error) {
	c, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %s", err.Error())
	}

	gitProvider := gitprovider.GetGitProvider(gitProviderId, c.GitProviders)
	if gitProvider == nil {
		return nil, fmt.Errorf("git provider not found")
	}

	response, err := gitProvider.GetNamespaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get namespaces: %s", err.Error())
	}

	return response, nil
}
