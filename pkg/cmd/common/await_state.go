// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"errors"
	"net/http"
	"time"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
)

func AwaitWorkspaceState(workspaceId string, stateName apiclient.ModelsResourceStateName) error {
	for {
		ws, _, err := apiclient_util.GetWorkspace(workspaceId, false)
		if err != nil {
			return err
		}
		if ws.State.Name == stateName {
			return nil
		}
		if ws.State.Name == apiclient.ResourceStateNameError {
			var errorMessage string
			if ws.State.Error != nil {
				errorMessage = *ws.State.Error
			}
			return errors.New(errorMessage)
		}
		time.Sleep(time.Second)
	}
}

func AwaitTargetState(targetId string, stateName apiclient.ModelsResourceStateName) error {
	for {
		t, _, err := apiclient_util.GetTarget(targetId, false)
		if err != nil {
			return err
		}
		if t.State.Name == stateName || t.TargetConfig.ProviderInfo.AgentlessTarget != nil && *t.TargetConfig.ProviderInfo.AgentlessTarget {
			return nil
		}
		if t.State.Name == apiclient.ResourceStateNameError {
			var errorMessage string
			if t.State.Error != nil {
				errorMessage = *t.State.Error
			}
			return errors.New(errorMessage)
		}
		time.Sleep(time.Second)
	}
}

func AwaitWorkspaceDeleted(workspaceId string) error {
	for {
		_, statusCode, err := apiclient_util.GetWorkspace(workspaceId, false)
		if err != nil {
			if statusCode == http.StatusNotFound {
				return nil
			}
			return err
		}
		time.Sleep(time.Second)
	}
}

func AwaitTargetDeleted(workspaceId string) error {
	for {
		_, statusCode, err := apiclient_util.GetTarget(workspaceId, false)
		if err != nil {
			if statusCode == http.StatusNotFound {
				return nil
			}
			return err
		}
		time.Sleep(time.Second)
	}
}
