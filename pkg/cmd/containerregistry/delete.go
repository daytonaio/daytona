// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package containerregistry

import (
	"context"
	"net/url"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/views"
	containerregistry_view "github.com/daytonaio/daytona/pkg/views/containerregistry"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/spf13/cobra"
)

var containerRegistryDeleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"remove", "rm"},
	Short:   "Delete a container registry",
	Args:    cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var registryDto *apiclient.ContainerRegistry
		var selectedServer string

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			c, err := config.GetConfig()
			if err != nil {
				return err
			}

			activeProfile, err := c.GetActiveProfile()
			if err != nil {
				return err
			}

			containerRegistries, res, err := apiClient.ContainerRegistryAPI.ListContainerRegistries(context.Background()).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(containerRegistries) == 0 {
				views_util.NotifyEmptyContainerRegistryList(false)
				return nil
			}

			registryDto, err = containerregistry_view.GetRegistryFromPrompt(containerRegistries, activeProfile.Name, false)
			if err != nil {
				if common.IsCtrlCAbort(err) {
					return nil
				} else {
					return err
				}
			}

			selectedServer = registryDto.Server
		} else {
			selectedServer = args[0]
		}

		res, err := apiClient.ContainerRegistryAPI.RemoveContainerRegistry(context.Background(), url.QueryEscape(selectedServer)).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		views.RenderInfoMessage("Container registry deleted successfully")
		return nil
	},
}
