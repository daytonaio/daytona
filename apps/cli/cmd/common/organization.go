// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"context"

	"github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/config"
	daytonaapiclient "github.com/daytonaio/daytona/daytonaapiclient"
)

func GetPersonalOrganizationId(profile config.Profile) (string, error) {
	apiClient, err := apiclient.GetApiClient(&profile, nil)
	if err != nil {
		return "", err
	}

	organizationList, res, err := apiClient.OrganizationsAPI.ListOrganizations(context.Background()).Execute()
	if err != nil {
		return "", apiclient.HandleErrorResponse(res, err)
	}

	for _, organization := range organizationList {
		if organization.Personal {
			return organization.Id, nil
		}
	}

	return "", nil
}

func GetActiveOrganizationName(apiClient *daytonaapiclient.APIClient, ctx context.Context) (string, error) {
	activeOrganizationId, err := config.GetActiveOrganizationId()
	if err != nil {
		return "", err
	}

	if activeOrganizationId == "" {
		return "", config.ErrNoActiveOrganization
	}

	activeOrganization, res, err := apiClient.OrganizationsAPI.GetOrganization(ctx, activeOrganizationId).Execute()
	if err != nil {
		return "", apiclient.HandleErrorResponse(res, err)
	}

	return activeOrganization.Name, nil
}
