// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agentmode

import (
	"context"

	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/agent/config"
	cmd_common "github.com/daytonaio/daytona/pkg/cmd/common"
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

		mode := config.ModeTarget
		if isWorkspaceAgentMode() {
			mode = config.ModeWorkspace
		}

		cfg, err := config.GetConfig(mode)
		if err != nil {
			return err
		}

		if isWorkspaceAgentMode() {
			return runWorkspaceLogs(ctx, cfg.Server.ApiUrl, cfg.Server.ApiKey)
		}

		return runTargetLogs(ctx, cfg.Server.ApiUrl, cfg.Server.ApiKey)
	},
}

func runWorkspaceLogs(ctx context.Context, serverUrl, serverApiKey string) error {
	workspace, _, err := apiclient_util.GetWorkspace(workspaceId)
	if err != nil {
		return err
	}

	cmd_common.ReadWorkspaceLogs(ctx, cmd_common.ReadLogParams{
		Id:        workspace.Id,
		Label:     &workspace.Name,
		ServerUrl: serverUrl,
		ApiKey:    serverApiKey,
		Index:     util.Pointer(0),
		Follow:    util.Pointer(false),
	})

	return nil
}

func runTargetLogs(ctx context.Context, serverUrl, serverApiKey string) error {
	target, _, err := apiclient_util.GetTarget(targetId)
	if err != nil {
		return err
	}

	cmd_common.ReadTargetLogs(ctx, cmd_common.ReadLogParams{
		Id:        target.Id,
		Label:     &target.Name,
		ServerUrl: serverUrl,
		ApiKey:    serverApiKey,
		Index:     util.Pointer(0),
		Follow:    util.Pointer(false),
	})

	return nil
}
