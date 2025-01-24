// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"context"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/common"
	runner "github.com/daytonaio/daytona/pkg/views/server/runner/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
)

func GetRunnerFlow(apiClient *apiclient.APIClient, action string) (*runner.RunnerView, error) {
	ctx := context.Background()

	c, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	activeProfile, err := c.GetActiveProfile()
	if err != nil {
		return nil, err
	}

	runners, res, err := apiClient.RunnerAPI.ListRunners(ctx).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	if len(runners) == 0 {
		views_util.NotifyEmptyRunnerList(true)
		return nil, nil
	}

	if len(runners) == 1 {
		return &runner.RunnerView{
			Id:   runners[0].Id,
			Name: runners[0].Name,
		}, nil
	}

	selectedRunner, err := runner.GetRunnerFromPrompt(runners, activeProfile.Name, action)
	if err != nil {
		if common.IsCtrlCAbort(err) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return selectedRunner, nil
}
