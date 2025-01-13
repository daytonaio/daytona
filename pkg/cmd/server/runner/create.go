// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"context"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/views/server/runner"
	"github.com/docker/docker/pkg/stringid"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a runner",
	Args:    cobra.NoArgs,
	Aliases: common.GetAliases("create"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		runnerList, res, err := apiClient.RunnerAPI.ListRunners(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		existingRunnerNames := util.ArrayMap(runnerList, func(r apiclient.RunnerDTO) string {
			return r.Name
		})

		name := nameFlag

		if name == "" {
			err = runner.RunnerCreationView(&name, existingRunnerNames)
			if err != nil {
				return err
			}
		}

		id := stringid.GenerateRandomID()
		id = stringid.TruncateID(id)

		runnerDto, res, err := apiClient.RunnerAPI.CreateRunner(ctx).Runner(apiclient.CreateRunnerDTO{
			Id:   id,
			Name: name,
		}).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		apiServerConfig, res, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		apiUrl := util.GetFrpcApiUrl(apiServerConfig.Frps.Protocol, apiServerConfig.Id, apiServerConfig.Frps.Domain)
		runner.Notify(runnerDto, apiUrl, c.Id, !c.TelemetryEnabled)

		return nil
	},
}

var nameFlag string

func init() {
	createCmd.Flags().StringVarP(&nameFlag, "name", "n", "", "Runner name")
}
