// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agentmode

import (
	"fmt"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	target_cmd "github.com/daytonaio/daytona/pkg/cmd/target"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:     "stop",
	Short:   "Stop the project",
	Args:    cobra.NoArgs,
	GroupID: util.TARGET_GROUP,
	RunE: func(cmd *cobra.Command, args []string) error {
		apiClient, err := apiclient.GetApiClient(nil)
		if err != nil {
			return err
		}

		err = target_cmd.StopTarget(apiClient, targetId, projectName)
		if err != nil {
			return err
		}

		if projectName != "" {
			views.RenderInfoMessage(fmt.Sprintf("Project '%s' from target '%s' successfully stopped", projectName, targetId))
		} else {
			views.RenderInfoMessage(fmt.Sprintf("Target '%s' successfully stopped", targetId))
		}
		return nil
	},
}
