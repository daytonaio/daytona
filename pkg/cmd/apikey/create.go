// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikey

import (
	"context"
	"errors"

	"github.com/spf13/cobra"

	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/common"
	view "github.com/daytonaio/daytona/pkg/views/apikey"
)

var createCmd = &cobra.Command{
	Use:     "create [NAME]",
	Short:   "Create a new API key",
	Args:    cobra.RangeArgs(0, 1),
	Aliases: common.GetAliases("create"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		var keyName string

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		apiKeyList, _, err := apiClient.ApiKeyAPI.ListClientApiKeys(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(nil, err)
		}

		if len(args) == 1 {
			keyName = args[0]
		} else {
			view.ApiKeyCreationView(&keyName, apiKeyList)
		}

		for _, key := range apiKeyList {
			if key.Name == keyName {
				return errors.New("key name already exists, please choose a different one")
			}
		}

		key, _, err := apiClient.ApiKeyAPI.CreateApiKey(ctx, keyName).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(nil, err)
		}

		serverConfig, _, err := apiClient.ServerAPI.GetConfigExecute(apiclient.ApiGetConfigRequest{})
		if err != nil {
			return err
		}

		if serverConfig.Frps == nil {
			return errors.New("frps config is missing")
		}

		apiUrl := util.GetFrpcApiUrl(serverConfig.Frps.Protocol, serverConfig.Id, serverConfig.Frps.Domain)

		view.Render(key, apiUrl)
		return nil
	},
}
