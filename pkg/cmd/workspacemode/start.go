// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspacemode

import (
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	workspace_cmd "github.com/daytonaio/daytona/pkg/cmd/workspace"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:     "start",
	Short:   "Start the project",
	Args:    cobra.NoArgs,
	GroupID: util.WORKSPACE_GROUP,
	RunE: func(cmd *cobra.Command, args []string) error {
		apiClient, err := apiclient.GetApiClient(nil)
		if err != nil {
			return err
		}

		err = workspace_cmd.StartWorkspace(apiClient, workspaceId, projectName)
		if err != nil {
			return err
		}

		views.RenderInfoMessage("Project successfully started")
		return nil
	},
}
