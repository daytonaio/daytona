// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspacemode

import (
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	"github.com/daytonaio/daytona/pkg/views/workspace/info"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:     "info",
	Short:   "Show project info",
	Aliases: []string{"view", "inspect"},
	Args:    cobra.ExactArgs(0),
	GroupID: util.WORKSPACE_GROUP,
	RunE: func(cmd *cobra.Command, args []string) error {
		var workspace *apiclient.WorkspaceDTO

		workspace, err := apiclient_util.GetWorkspace(workspaceId, true)
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

		info.Render(workspace, "", false)
		return nil
	},
}

func init() {
	format.RegisterFormatFlag(infoCmd)
}
