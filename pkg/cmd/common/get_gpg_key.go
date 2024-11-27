// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"context"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
)

func GetGitProviderGpgKey(apiClient *apiclient.APIClient, ctx context.Context, providerConfigId *string) (string, error) {
	if providerConfigId == nil || *providerConfigId == "" {
		return "", nil
	}

	var gpgKey string

	gitProvider, res, err := apiClient.GitProviderAPI.GetGitProvider(ctx, *providerConfigId).Execute()
	if err != nil {
		return "", apiclient_util.HandleErrorResponse(res, err)
	}

	// Extract GPG key if present
	if gitProvider != nil {
		if gitProvider.SigningMethod != nil && gitProvider.SigningKey != nil {
			if *gitProvider.SigningMethod == apiclient.SigningMethodGPG {
				gpgKey = *gitProvider.SigningKey
			}
		}
	}

	return gpgKey, nil
}
