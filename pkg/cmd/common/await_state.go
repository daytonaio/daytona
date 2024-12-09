// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"context"
	"errors"
	"net/http"
	"time"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
)

func AwaitWorkspaceState(workspaceId string, stateName apiclient.ModelsResourceStateName) error {
	for {
		ws, _, err := apiclient_util.GetWorkspace(workspaceId)
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
		t, _, err := apiclient_util.GetTarget(targetId)
		if err != nil {
			return err
		}

		if t.State.Name == stateName || t.State.Name == apiclient.ResourceStateNameUndefined {
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

func AwaitProviderInstalled(runnerId, providerName, version string) error {
	for {
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		runner, res, err := apiClient.RunnerAPI.GetRunner(ctx, runnerId).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		if runner.Metadata == nil {
			continue
		}

		for _, provider := range runner.Metadata.Providers {
			if provider.Name == providerName && provider.Version == version {
				return nil
			}
		}

		time.Sleep(time.Second)
	}
}

func AwaitWorkspaceDeleted(workspaceId string) error {
	for {
		_, statusCode, err := apiclient_util.GetWorkspace(workspaceId)
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
		_, statusCode, err := apiclient_util.GetTarget(workspaceId)
		if err != nil {
			if statusCode == http.StatusNotFound {
				return nil
			}
			return err
		}
		time.Sleep(time.Second)
	}
}
