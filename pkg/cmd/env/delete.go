// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package env

import (
	"context"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/env"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:     "remove [KEY]...",
	Short:   "Remove server environment variables",
	Aliases: []string{"r", "rm", "d", "delete"},
	RunE: func(cmd *cobra.Command, args []string) error {
		keys := []string{}

		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if len(args) > 0 {
			keys = args
		} else {
			selectedEnvVars, err := env.RemoveEnvVarsView(ctx, *apiClient)
			if err != nil {
				return err
			}

			for _, envVar := range selectedEnvVars {
				keys = append(keys, envVar.Key)
			}
		}

		for _, key := range keys {
			res, err := apiClient.EnvVarAPI.DeleteEnvironmentVariable(ctx, key).Execute()
			if err != nil {
				log.Error(apiclient_util.HandleErrorResponse(res, err))
			}
		}

		views.RenderInfoMessageBold("Server environment variables have been successfully removed")

		return nil
	},
}
