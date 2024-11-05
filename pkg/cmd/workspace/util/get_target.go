// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

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
	Ctx               context.Context
	ApiClient         *apiclient.APIClient
	ActiveProfileName string
	TargetNameFlag    string
	PromptUsingTUI    bool
}

func GetTarget(params GetTargetConfigParams) (t *apiclient.TargetDTO, targetExisted bool, err error) {
	targetList, res, err := params.ApiClient.TargetAPI.ListTargets(params.Ctx).Execute()
	if err != nil {
		return nil, false, apiclient_util.HandleErrorResponse(res, err)
	}

	if params.TargetNameFlag != "" {
		for _, t := range targetList {
			if t.Name == params.TargetNameFlag {
				return &t, true, nil
			}
		}
		return nil, false, fmt.Errorf("target config '%s' not found", params.TargetNameFlag)
	}

	if !params.PromptUsingTUI {
		for _, t := range targetList {
			if t.Default {
				return &t, true, nil
			}
		}
	}

	if len(targetList) == 0 {
		t, err := runCreateTargetDtoFlow(params)
		if err != nil {
			return nil, false, err
		}
		return t, true, nil
	}

	selectedTarget := selection.GetTargetFromPrompt(targetList, true, "Use")

	if selectedTarget == nil {
		return nil, false, common.ErrCtrlCAbort
	}

	if selectedTarget.Name == selection.NewTargetIdentifier {
		t, err := runCreateTargetDtoFlow(params)
		if err != nil {
			return nil, false, err
		}
		return t, true, nil
	}

	return selectedTarget, true, nil
}

func runCreateTargetDtoFlow(params GetTargetConfigParams) (*apiclient.TargetDTO, error) {
	createTargetDto, err := target.CreateTargetDtoFlow(target.TargetCreationParams{
		Ctx:               params.Ctx,
		ApiClient:         params.ApiClient,
		ActiveProfileName: params.ActiveProfileName,
	})
	if err != nil {
		return nil, err
	}

	return &apiclient.TargetDTO{
		Name:    createTargetDto.Name,
		Options: createTargetDto.Options,
		ProviderInfo: apiclient.TargetProviderInfo{
			Name:    createTargetDto.ProviderInfo.Name,
			Version: createTargetDto.ProviderInfo.Version,
		},
	}, nil
}
