// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package env

import (
	"context"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var envVars []string

var setCmd = &cobra.Command{
	Use:     "set",
	Short:   "Set profile environment variables",
	Aliases: []string{"s", "update", "add", "delete", "rm"},
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

		if profileData.EnvVars == nil {
			profileData.EnvVars = &map[string]string{}
		}

		form := huh.NewForm(
			huh.NewGroup(
				views.GetEnvVarsInput(profileData.EnvVars),
			),
		).WithTheme(views.GetCustomTheme())

		err = form.Run()
		if err != nil {
			log.Fatal(err)
		}

		res, err = apiClient.ProfileAPI.SetProfileData(ctx).ProfileData(*profileData).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		views.RenderInfoMessageBold("Profile environment variables have been successfully set")
	},
}

func init() {
	setCmd.Flags().StringArrayP("var", "e", envVars, "Environment variables to set")
}
