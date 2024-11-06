// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"context"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/spf13/cobra"
)

type workspaceState string

const (
	WORKSPACE_STATE_RUNNING workspaceState = "Running"
	WORKSPACE_STATE_STOPPED workspaceState = "Unavailable"
)

func GetWorkspaceNameCompletions() ([]string, cobra.ShellCompDirective) {
	ctx := context.Background()
	apiClient, err := apiclient_util.GetApiClient(nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	workspaceList, _, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var choices []string
	for _, v := range workspaceList {
		choices = append(choices, v.Name)
	}

	return choices, cobra.ShellCompDirectiveNoFileComp
}

func GetAllWorkspacesByState(state workspaceState) ([]string, cobra.ShellCompDirective) {
	ctx := context.Background()
	apiClient, err := apiclient_util.GetApiClient(nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	workspaceList, _, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var choices []string
	for _, workspace := range workspaceList {
		if state == WORKSPACE_STATE_RUNNING && workspace.State.Uptime != 0 {
			choices = append(choices, workspace.Name)
			break
		}
		if state == WORKSPACE_STATE_STOPPED && workspace.State.Uptime == 0 {
			choices = append(choices, workspace.Name)
			break
		}
	}

	return choices, cobra.ShellCompDirectiveNoFileComp
}
