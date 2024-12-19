// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"context"

	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views/server/runner"
	runner_view "github.com/daytonaio/daytona/pkg/views/server/runner"
	"github.com/docker/docker/pkg/stringid"

	"github.com/spf13/cobra"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register runner",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

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
			err = runner.RunnerRegistrationView(&name, existingRunnerNames)
			if err != nil {
				return err
			}
		}

		id := stringid.GenerateRandomID()
		id = stringid.TruncateID(id)

		runner, res, err := apiClient.RunnerAPI.RegisterRunner(ctx).Runner(apiclient.RegisterRunnerDTO{
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
		runner_view.Notify(runner, apiUrl)

		return nil
	},
}

var nameFlag string

func init() {
	registerCmd.Flags().StringVarP(&nameFlag, "name", "n", "", "Runner name")
}
