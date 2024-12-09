// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package conversion

import (
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/models"
)

func ToGitProviderConfig(gitProviderConfig *apiclient.GitProvider) *models.GitProviderConfig {
	if gitProviderConfig == nil {
		return nil
	}

	return &models.GitProviderConfig{
		Id:            gitProviderConfig.Id,
		ProviderId:    gitProviderConfig.ProviderId,
		Username:      gitProviderConfig.Username,
		BaseApiUrl:    gitProviderConfig.BaseApiUrl,
		Token:         gitProviderConfig.Token,
		Alias:         gitProviderConfig.Alias,
		SigningKey:    gitProviderConfig.SigningKey,
		SigningMethod: util.Pointer(models.SigningMethod(*gitProviderConfig.SigningMethod)),
	}
}
