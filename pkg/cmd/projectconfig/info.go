// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package projectconfig

import (
	"context"
	"net/http"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	"github.com/daytonaio/daytona/pkg/views/projectconfig/info"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var projectConfigInfoCmd = &cobra.Command{
	Use:     "info",
	Short:   "Show project config info",
	Aliases: []string{"view", "inspect"},
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		apiServerConfig, res, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		var projectConfig *apiclient.ProjectConfig

		if len(args) == 0 {
			projectConfigList, res, err := apiClient.ProjectConfigAPI.ListProjectConfigs(ctx).Execute()
			if err != nil {
				log.Fatal(apiclient_util.HandleErrorResponse(res, err))
			}

			if format.FormatFlag != "" {
				format.UnblockStdOut()
			}

			projectConfig = selection.GetProjectConfigFromPrompt(projectConfigList, 0, false, false, "View")
			if format.FormatFlag != "" {
				format.BlockStdOut()
			}

		} else {
			var res *http.Response
			projectConfig, res, err = apiClient.ProjectConfigAPI.GetProjectConfig(ctx, args[0]).Execute()
			if err != nil {
				log.Fatal(apiclient_util.HandleErrorResponse(res, err))
			}
		}

		if projectConfig == nil {
			return
		}

		if format.FormatFlag != "" {
			formattedData := format.NewFormatter(projectConfig)
			formattedData.Print()
			return
		}

		info.Render(projectConfig, apiServerConfig, false)
	},
}

func init() {
	format.RegisterFormatFlag(projectConfigInfoCmd)
}
