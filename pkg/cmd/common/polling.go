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
	ctx := context.Background()

	apiClient, err := apiclient_util.GetApiClient(nil)
	if err != nil {
		return err
	}

	for {
		state, _, err := apiClient.WorkspaceAPI.GetWorkspaceState(ctx, workspaceId).Execute()
		if err != nil {
			return err
		}
		if state.Name == stateName {
			return nil
		}
		if state.Name == apiclient.ResourceStateNameError {
			var errorMessage string
			if state.Error != nil {
				errorMessage = *state.Error
			}
			return errors.New(errorMessage)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func AwaitTargetState(targetId string, stateName apiclient.ModelsResourceStateName) error {
	ctx := context.Background()

	apiClient, err := apiclient_util.GetApiClient(nil)
	if err != nil {
		return err
	}

	for {
		state, _, err := apiClient.TargetAPI.GetTargetState(ctx, targetId).Execute()
		if err != nil {
			return err
		}

		if state.Name == stateName || state.Name == apiclient.ResourceStateNameUndefined {
			return nil
		}

		if state.Name == apiclient.ResourceStateNameError {
			var errorMessage string
			if state.Error != nil {
				errorMessage = *state.Error
			}
			return errors.New(errorMessage)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func AwaitProviderInstalled(runnerId, providerName, version string) error {
	ctx := context.Background()

	apiClient, err := apiclient_util.GetApiClient(nil)
	if err != nil {
		return err
	}

	for {
		providers, res, err := apiClient.ProviderAPI.GetRunnerProviders(ctx, runnerId).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		if providers == nil {
			continue
		}

		for _, provider := range providers {
			if provider.Name == providerName && provider.Version == version {
				return nil
			}
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func AwaitWorkspaceDeleted(workspaceId string) error {
	ctx := context.Background()

	apiClient, err := apiclient_util.GetApiClient(nil)
	if err != nil {
		return err
	}

	for {
		_, res, err := apiClient.WorkspaceAPI.GetWorkspaceState(ctx, workspaceId).Execute()
		if err != nil {
			if res != nil && res.StatusCode == http.StatusNotFound {
				return nil
			}
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func AwaitTargetDeleted(workspaceId string) error {
	ctx := context.Background()

	apiClient, err := apiclient_util.GetApiClient(nil)
	if err != nil {
		return err
	}

	for {
		_, res, err := apiClient.TargetAPI.GetTargetState(ctx, workspaceId).Execute()
		if err != nil {
			if res != nil && res.StatusCode == http.StatusNotFound {
				return nil
			}
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}
}
