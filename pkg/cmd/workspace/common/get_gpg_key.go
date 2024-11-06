// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"context"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/gitprovider"
)

func GetGitProviderGpgKey(apiClient *apiclient.APIClient, ctx context.Context, providerConfigId *string) (string, error) {
	if providerConfigId == nil || *providerConfigId == "" {
		return "", nil
	}

	var providerConfig *gitprovider.GitProviderConfig
	var gpgKey string

	gitProvider, res, err := apiClient.GitProviderAPI.GetGitProvider(ctx, *providerConfigId).Execute()
	if err != nil {
		return "", apiclient_util.HandleErrorResponse(res, err)
	}

	// Extract GPG key if present
	if gitProvider != nil {
		providerConfig = &gitprovider.GitProviderConfig{
			SigningMethod: (*gitprovider.SigningMethod)(gitProvider.SigningMethod),
			SigningKey:    gitProvider.SigningKey,
		}

		if providerConfig.SigningMethod != nil && providerConfig.SigningKey != nil {
			if *providerConfig.SigningMethod == gitprovider.SigningMethodGPG {
				gpgKey = *providerConfig.SigningKey
			}
		}
	}

	return gpgKey, nil
}
