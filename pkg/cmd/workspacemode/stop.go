// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspacemode

import (
	"fmt"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	workspace_cmd "github.com/daytonaio/daytona/pkg/cmd/workspace"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:     "stop",
	Short:   "Stop the project",
	Args:    cobra.NoArgs,
	GroupID: util.WORKSPACE_GROUP,
	RunE: func(cmd *cobra.Command, args []string) error {
		apiClient, err := apiclient.GetApiClient(nil)
		if err != nil {
			return err
		}

		err = workspace_cmd.StopWorkspace(apiClient, workspaceId, projectName)
		if err != nil {
			return err
		}

		if projectName != "" {
			views.RenderInfoMessage(fmt.Sprintf("Project '%s' from workspace '%s' successfully stopped", projectName, workspaceId))
		} else {
			views.RenderInfoMessage(fmt.Sprintf("Workspace '%s' successfully stopped", workspaceId))
		}
		return nil
	},
}
