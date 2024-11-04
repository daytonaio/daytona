// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agentmode

import (
	"errors"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:     "start",
	Short:   "Start the workspace",
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

	// 	err = target_cmd.StartTarget(apiClient, targetId, workspaceId)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	views.RenderInfoMessage("Workspace successfully started")
	// 	return nil
	// },
}
