// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package env

import (
	"context"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	"github.com/daytonaio/daytona/pkg/views/env"
	"github.com/spf13/cobra"
)

var showValuesFlag bool

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List server environment variables",
	Aliases: common.GetAliases("list"),
	RunE: func(cmd *cobra.Command, args []string) error {
		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}
		ctx := context.Background()

		envVars, res, err := apiClient.EnvVarAPI.ListEnvironmentVariables(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		if !showValuesFlag {
			for i := range envVars {
				envVars[i].Value = "****************"
			}
		}

		if format.FormatFlag != "" {
			if envVars == nil {
				envVars = []apiclient.EnvironmentVariable{}
			}

			formattedData := format.NewFormatter(envVars)
			formattedData.Print()
			return nil
		}

		env.List(envVars)
		return nil
	},
}

func init() {
	format.RegisterFormatFlag(listCmd)

	listCmd.Flags().BoolVarP(&showValuesFlag, "show-values", "v", false, "Show environment variable values")
}
