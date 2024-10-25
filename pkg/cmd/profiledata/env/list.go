// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package env

import (
	"context"

	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	"github.com/daytonaio/daytona/pkg/views/env"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List profile environment variables",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		apiClient, err := apiclient.GetApiClient(nil)
		if err != nil {
			return err
		}
		ctx := context.Background()

		profileData, res, err := apiClient.ProfileAPI.GetProfileData(ctx).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		if format.FormatFlag != "" {
			if profileData.EnvVars == nil {
				profileData.EnvVars = map[string]string{}
			}
			formattedData := format.NewFormatter(profileData.EnvVars)
			formattedData.Print()
			return nil
		}

		env.List(profileData.EnvVars)
		return nil
	},
}

func init() {
	format.RegisterFormatFlag(listCmd)
}
