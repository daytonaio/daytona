// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/server/runner"
	"github.com/docker/docker/pkg/stringid"

	"github.com/spf13/cobra"
)

var runnerRegisterCmd = &cobra.Command{
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
			return r.Alias
		})

		alias := runnerAliasFlag

		if alias == "" {
			err = runner.RunnerRegistrationView(&alias, existingRunnerNames)
			if err != nil {
				return err
			}
		}

		id := stringid.GenerateRandomID()
		id = stringid.TruncateID(id)

		runner, res, err := apiClient.RunnerAPI.RegisterRunner(ctx).Runner(apiclient.RegisterRunnerDTO{
			Id:    id,
			Alias: alias,
		}).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		views.RenderInfoMessage(fmt.Sprintf("Runner %s registered successfully", runner.Alias))
		return nil
	},
}

var runnerAliasFlag string

func init() {
	runnerRegisterCmd.Flags().StringVarP(&runnerAliasFlag, "alias", "a", "", "Runner alias")
}
