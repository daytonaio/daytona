// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"context"
	"fmt"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/target"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/views/target/selection"
)

type GetTargetConfigParams struct {
	ApiClient         *apiclient.APIClient
	ActiveProfileName string
	TargetNameFlag    string
	PromptUsingTUI    bool
}

func GetTarget(ctx context.Context, params GetTargetConfigParams) (t *apiclient.TargetDTO, createTargetDto *apiclient.CreateTargetDTO, err error) {
	targetList, res, err := params.ApiClient.TargetAPI.ListTargets(ctx).Execute()
	if err != nil {
		return nil, nil, apiclient_util.HandleErrorResponse(res, err)
	}

	if params.TargetNameFlag != "" {
		for _, t := range targetList {
			if t.Name == params.TargetNameFlag {
				return &t, nil, nil
			}
		}
		return nil, nil, fmt.Errorf("target '%s' not found", params.TargetNameFlag)
	}

	if !params.PromptUsingTUI {
		for _, t := range targetList {
			if t.Default {
				return &t, nil, nil
			}
		}
	}

	if len(targetList) == 0 {
		createTargetDto, err := target.CreateTargetDtoFlow(ctx, target.TargetCreationParams{
			ApiClient:         params.ApiClient,
			ActiveProfileName: params.ActiveProfileName,
		})
		if err != nil {
			return nil, nil, err
		}
		return nil, createTargetDto, nil
	}

	selectedTarget := selection.GetTargetFromPrompt(targetList, true, "Use")

	if selectedTarget == nil {
		return nil, nil, common.ErrCtrlCAbort
	}

	if selectedTarget.Name == selection.NewTargetIdentifier {
		createTargetDto, err := target.CreateTargetDtoFlow(ctx, target.TargetCreationParams{
			ApiClient:         params.ApiClient,
			ActiveProfileName: params.ActiveProfileName,
		})
		if err != nil {
			return nil, createTargetDto, err
		}
		return nil, createTargetDto, nil
	}

	return selectedTarget, nil, nil
}
