// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agentmode

import (
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	target_views "github.com/daytonaio/daytona/pkg/views/target/info"
	workspaces_views "github.com/daytonaio/daytona/pkg/views/workspace/info"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:     "info",
	Short:   "Show resource info",
	Aliases: []string{"view", "inspect"},
	Args:    cobra.ExactArgs(0),
	GroupID: util.TARGET_GROUP,
	RunE: func(cmd *cobra.Command, args []string) error {
		if isWorkspaceAgentMode() {
			return runWorkspaceInfo()
		}

		return runTargetInfo()
	},
}

func init() {
	format.RegisterFormatFlag(infoCmd)
}

func runWorkspaceInfo() error {
	workspace, _, err := apiclient_util.GetWorkspace(workspaceId, true)
	if err != nil {
		return err
	}

	if workspace == nil {
		return nil
	}

	if format.FormatFlag != "" {
		formattedData := format.NewFormatter(workspace)
		formattedData.Print()
		return nil
	}

	workspaces_views.Render(workspace, "", false)

	return nil
}

func runTargetInfo() error {
	target, _, err := apiclient_util.GetTarget(targetId, true)
	if err != nil {
		return err
	}

	if target == nil {
		return nil
	}

	if format.FormatFlag != "" {
		formattedData := format.NewFormatter(target)
		formattedData.Print()
		return nil
	}

	target_views.Render(target, false)

	return nil
}
