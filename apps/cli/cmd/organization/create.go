// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package organization

import (
	"context"

	"github.com/daytonaio/apiclient"
	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/cli/views/common"
	"github.com/daytonaio/daytona/cli/views/organization"
	"github.com/spf13/cobra"
)

var CreateCmd = &cobra.Command{
	Use:   "create [ORGANIZATION_NAME]",
	Short: "Create a new organization and set it as active",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient_cli.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		createOrganizationDto := apiclient.CreateOrganization{
			Name: args[0],
		}

		org, res, err := apiClient.OrganizationsAPI.CreateOrganization(ctx).CreateOrganization(createOrganizationDto).Execute()
		if err != nil {
			return apiclient_cli.HandleErrorResponse(res, err)
		}

		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			return err
		}

		activeProfile.ActiveOrganizationId = &org.Id
		err = c.EditProfile(activeProfile)
		if err != nil {
			return err
		}

		organization.RenderInfo(org, false)

		common.RenderInfoMessageBold("Your organization has been created and its approval is pending\nOur team has been notified and will set up your resource quotas shortly")
		return nil
	},
}
