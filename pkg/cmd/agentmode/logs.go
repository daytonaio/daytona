// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agentmode

import (
	"context"

	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:     "logs",
	Short:   "View resource logs",
	Args:    cobra.NoArgs,
	GroupID: util.TARGET_GROUP,
	Aliases: []string{"lg", "log"},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		err := runTargetLogs(ctx)
		if err != nil {
			return err
		}

		if isWorkspaceAgentMode() {
			return runWorkspaceLogs(ctx)
		}

		return nil
	},
}

func runWorkspaceLogs(ctx context.Context) error {
	workspace, _, err := apiclient_util.GetWorkspace(workspaceId, false)
	if err != nil {
		return err
	}

	apiclient_util.ReadWorkspaceLogs(ctx, apiclient_util.ReadLogParams{
		Id:    workspace.Id,
		Label: &workspace.Name,
	})

	return nil
}

func runTargetLogs(ctx context.Context) error {
	target, _, err := apiclient_util.GetTarget(targetId, false)
	if err != nil {
		return err
	}

	apiclient_util.ReadTargetLogs(ctx, apiclient_util.ReadLogParams{
		Id:    target.Id,
		Label: &target.Name,
	})

	return nil
}
