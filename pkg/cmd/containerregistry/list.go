// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package containerregistry

import (
	"context"

	"github.com/daytonaio/daytona/internal/util/apiclient"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	containerregistry_view "github.com/daytonaio/daytona/pkg/views/containerregistry/list"
	"github.com/spf13/cobra"
)

var containerRegistryListCmd = &cobra.Command{
	Use:     "list",
	Short:   "Lists container registries",
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		containerRegistries, res, err := apiClient.ContainerRegistryAPI.ListContainerRegistries(context.Background()).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		if format.FormatFlag != "" {
			formattedData := format.NewFormatter(containerRegistries)
			formattedData.Print()
			return nil
		}

		containerregistry_view.ListRegistries(containerRegistries)
		return nil
	},
}

func init() {
	format.RegisterFormatFlag(containerRegistryListCmd)
}
