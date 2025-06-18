// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/cli/views/common"
	"github.com/daytonaio/daytona/cli/views/organization"
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

func SelectOrganizationForFirstTimeLogin(profile config.Profile) (string, error) {
	apiClient, err := apiclient.GetApiClient(&profile, nil)
	if err != nil {
		return "", err
	}

	organizationList, res, err := apiClient.OrganizationsAPI.ListOrganizations(context.Background()).Execute()
	if err != nil {
		return "", apiclient.HandleErrorResponse(res, err)
	}

	if len(organizationList) == 0 {
		return "", fmt.Errorf("no organizations found")
	}

	// If there's only one organization, use it automatically
	if len(organizationList) == 1 {
		common.RenderInfoMessageBold(fmt.Sprintf("Using organization: %s", organizationList[0].Name))
		return organizationList[0].Id, nil
	}

	// Multiple organizations available - prompt user to select
	common.RenderInfoMessage("Multiple organizations available. Please select one:")

	chosenOrganization, err := organization.GetOrganizationIdFromPrompt(organizationList)
	if err != nil {
		return "", err
	}

	if chosenOrganization == nil {
		return "", fmt.Errorf("no organization selected")
	}

	common.RenderInfoMessageBold(fmt.Sprintf("Selected organization: %s", chosenOrganization.Name))
	return chosenOrganization.Id, nil
}

// ValidateActiveOrganization checks if the given organization ID is still available to the user
func ValidateActiveOrganization(profile config.Profile, organizationId string) (bool, error) {
	apiClient, err := apiclient.GetApiClient(&profile, nil)
	if err != nil {
		return false, err
	}

	organizationList, res, err := apiClient.OrganizationsAPI.ListOrganizations(context.Background()).Execute()
	if err != nil {
		return false, apiclient.HandleErrorResponse(res, err)
	}

	// Check if the organization ID exists in the user's available organizations
	for _, org := range organizationList {
		if org.Id == organizationId {
			return true, nil
		}
	}

	return false, nil
}
