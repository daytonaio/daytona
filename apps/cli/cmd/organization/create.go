// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package organization

import (
	"context"
	"fmt"

	"github.com/charmbracelet/huh"
	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/cli/views/common"
	"github.com/daytonaio/daytona/cli/views/organization"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/spf13/cobra"
)

var regionFlag string

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

		regions, res, err := apiClient.RegionsAPI.ListSharedRegions(ctx).Execute()
		if err != nil {
			return apiclient_cli.HandleErrorResponse(res, err)
		}

		var chosenRegion *apiclient.Region
		switch {
		case regionFlag != "":
			chosenRegion, err = resolveRegion(regions, regionFlag)
			if err != nil {
				return err
			}
		case len(regions) == 0:
			return fmt.Errorf("no shared regions available; contact your administrator")
		case len(regions) == 1:
			chosenRegion = &regions[0]
		default:
			var chosenRegionId string
			var regionOptions []huh.Option[string]

			for _, region := range regions {
				regionOptions = append(regionOptions, huh.NewOption(region.Name, region.Id))
			}

			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("Choose a Region").
						Options(
							regionOptions...,
						).
						Value(&chosenRegionId),
				).WithTheme(common.GetCustomTheme()),
			)

			if err := form.Run(); err != nil {
				return err
			}

			chosenRegion, err = resolveRegion(regions, chosenRegionId)
			if err != nil {
				return err
			}
		}

		createOrganizationDto := apiclient.CreateOrganization{
			Name:            args[0],
			DefaultRegionId: chosenRegion.Id,
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

// resolveRegion picks a region from the list whose Id or Name matches identifier.
// Pure function  no I/O.
func resolveRegion(regions []apiclient.Region, identifier string) (*apiclient.Region, error) {
	for _, region := range regions {
		if region.Id == identifier || region.Name == identifier {
			return &region, nil
		}
	}
	return nil, fmt.Errorf("region %q not found", identifier)
}

func init() {
	CreateCmd.Flags().StringVarP(&regionFlag, "region", "r", "", "Default region (id or name) for the organization")
}
