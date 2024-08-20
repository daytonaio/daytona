// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspacemode

import (
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	"github.com/daytonaio/daytona/pkg/views/workspace/info"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:     "info",
	Short:   "Show project info",
	Aliases: []string{"view", "inspect"},
	Args:    cobra.ExactArgs(0),
	GroupID: util.WORKSPACE_GROUP,
	Run: func(cmd *cobra.Command, args []string) {
		var workspace *apiclient.WorkspaceDTO

		workspace, err := apiclient_util.GetWorkspace(workspaceId)
		if err != nil {
			log.Fatal(err)
		}

		if workspace == nil {
			return
		}

		if format.FormatFlag != "" {
			formatter := format.NewFormatter(workspace)
			formatter.Print()
			return
		}

		info.Render(workspace, "", false)
	},
}

func init() {
	format.RegisterFormatFlag(infoCmd)
}
