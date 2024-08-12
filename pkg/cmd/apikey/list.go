// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikey

import (
	"context"

	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/output"
	"github.com/daytonaio/daytona/pkg/views/server/apikey"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var formatFlag string
var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List API keys",
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		apiClient, err := apiclient.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		apiKeyList, _, err := apiClient.ApiKeyAPI.ListClientApiKeys(ctx).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(nil, err))
		}

		if formatFlag != "" {
			display := output.NewOutputFormatter(apiKeyList, formatFlag)
			display.Print()
			return
		}

		apikey.ListApiKeys(apiKeyList)
	},
}

func init() {
	listCmd.PersistentFlags().StringVarP(&formatFlag, output.FormatFlagName, output.FormatFlagShortHand, formatFlag, output.FormatDescription)
	listCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if formatFlag != "" {
			output.BlockStdOut()
		}
	}
}
