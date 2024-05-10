// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package env

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List profile environment variables",
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		apiClient, err := server.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}
		ctx := context.Background()

		profileData, res, err := apiClient.ProfileAPI.GetProfileData(ctx).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		if profileData.EnvVars == nil || len(*profileData.EnvVars) == 0 {
			views.RenderInfoMessageBold("No environment variables set")
			return
		}

		for key, value := range *profileData.EnvVars {
			fmt.Printf("%s=%s\n", key, value)
		}
	},
}
