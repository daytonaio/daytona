// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package env

import (
	"context"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var setCmd = &cobra.Command{
	Use:     "set [KEY=VALUE]...",
	Short:   "Set profile environment variables",
	Aliases: []string{"s", "update", "add", "delete", "rm"},
	Run: func(cmd *cobra.Command, args []string) {
		apiClient, err := apiclient.GetApiClient(nil)
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

		if len(args) > 0 {
			for _, arg := range args {
				kv := strings.Split(arg, "=")
				if len(kv) != 2 {
					log.Fatalf("Invalid key-value pair: %s", arg)
				}
				(*profileData.EnvVars)[kv[0]] = kv[1]
			}
		} else {
			form := huh.NewForm(
				huh.NewGroup(
					views.GetEnvVarsInput(profileData.EnvVars),
				),
			).WithTheme(views.GetCustomTheme())

			err = form.Run()
			if err != nil {
				log.Fatal(err)
			}
		}

		res, err = apiClient.ProfileAPI.SetProfileData(ctx).ProfileData(*profileData).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		views.RenderInfoMessageBold("Profile environment variables have been successfully set")
	},
}
