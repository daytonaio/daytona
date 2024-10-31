// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
)

type TargetCreationParams struct {
	Ctx               context.Context
	ApiClient         *apiclient.APIClient
	ActiveProfileName string
}

func TargetCreationFlow(params TargetCreationParams) (*apiclient.TargetDTO, error) {
	targetList, res, err := params.ApiClient.TargetAPI.ListTargets(params.Ctx).Verbose(true).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	if len(targetList) == 0 {
		views_util.NotifyEmptyTargetList(true)
		return nil, nil
	}

	if format.FormatFlag != "" {
		format.UnblockStdOut()
	}

	// target = selection.GetTargetFromPrompt(targetList, false, "View")
	// if format.FormatFlag != "" {
	// 	format.BlockStdOut()
	// }

	return nil, nil
}
