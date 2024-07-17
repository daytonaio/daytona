// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikey

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views/server/apikey"
	view "github.com/daytonaio/daytona/pkg/views/server/apikey"
)

var GenerateCmd = &cobra.Command{
	Use:     "generate [NAME]",
	Short:   "Generate a new API key",
	Aliases: []string{"g", "new"},
	Args:    cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		var keyName string

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		apiKeyList, _, err := apiClient.ApiKeyAPI.ListClientApiKeys(ctx).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(nil, err))
		}

		if len(args) == 1 {
			keyName = args[0]
		} else {
			apikey.ApiKeyCreationView(&keyName, apiKeyList)
		}

		for _, key := range apiKeyList {
			if *key.Name == keyName {
				log.Fatal("key name already exists, please choose a different one")
			}
		}

		key, _, err := apiClient.ApiKeyAPI.GenerateApiKey(ctx, keyName).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(nil, err))
		}

		serverConfig, _, err := apiClient.ServerAPI.GetConfigExecute(apiclient.ApiGetConfigRequest{})
		if err != nil {
			log.Fatal(err)
		}

		apiUrl := util.GetFrpcApiUrl(*serverConfig.Frps.Protocol, *serverConfig.Id, *serverConfig.Frps.Domain)

		view.Render(key, apiUrl)
	},
}
