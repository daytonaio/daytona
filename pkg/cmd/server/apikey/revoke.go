// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikey

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/views/server/apikey"
	"github.com/daytonaio/daytona/pkg/views/util"
)

var revokeCmd = &cobra.Command{
	Use:     "revoke [NAME]",
	Short:   "Revoke an API key",
	Aliases: []string{"r", "rm", "delete"},
	Args:    cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		var keyName string

		apiClient, err := server.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		if len(args) == 1 {
			keyName = args[0]
		} else {
			apiKeyList, _, err := apiClient.ApiKeyAPI.ListClientApiKeys(ctx).Execute()
			if err != nil {
				log.Fatal(apiclient.HandleErrorResponse(nil, err))
			}

			keyName = apikey.GetApiKeyNameFromPrompt(apiKeyList, "Select an API key to revoke")
			if keyName == "" {
				log.Fatal("No API key selected")
			}
		}

		_, err = apiClient.ApiKeyAPI.RevokeApiKey(ctx, keyName).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(nil, err))
		}

		util.RenderInfoMessage("API key revoked")
	},
}
