// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agentmode

import (
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	target_cmd "github.com/daytonaio/daytona/pkg/cmd/target"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:     "start",
	Short:   "Start the workspace",
	Args:    cobra.NoArgs,
	GroupID: util.TARGET_GROUP,
	RunE: func(cmd *cobra.Command, args []string) error {
		apiClient, err := apiclient.GetApiClient(nil)
		if err != nil {
			return err
		}

		err = target_cmd.StartTarget(apiClient, targetId, workspaceName)
		if err != nil {
			return err
		}

		views.RenderInfoMessage("Workspace successfully started")
		return nil
	},
}
