// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agentmode

import (
	"errors"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:     "stop",
	Short:   "Stop the workspace",
	Args:    cobra.NoArgs,
	GroupID: util.TARGET_GROUP,
	RunE: func(cmd *cobra.Command, args []string) error {
		return errors.New("not implemented")
	},
	// RunE: func(cmd *cobra.Command, args []string) error {
	// 	apiClient, err := apiclient.GetApiClient(nil)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	err = target_cmd.StopTarget(apiClient, targetId, workspaceId)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	if workspaceId != "" {
	// 		views.RenderInfoMessage(fmt.Sprintf("Workspace '%s' from target '%s' successfully stopped", workspaceId, targetId))
	// 	} else {
	// 		views.RenderInfoMessage(fmt.Sprintf("Target '%s' successfully stopped", targetId))
	// 	}
	// 	return nil
	// },
}
