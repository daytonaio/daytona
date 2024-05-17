// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspacemode

import (
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/output"
	"github.com/daytonaio/daytona/pkg/views/workspace/info"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:     "info",
	Short:   "Show project info",
	Aliases: []string{"view"},
	Args:    cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		var workspace *apiclient.WorkspaceDTO

		workspace, err := apiclient_util.GetWorkspace(workspaceId)
		if err != nil {
			log.Fatal(err)
		}

		if workspace == nil {
			return
		}

		if output.FormatFlag != "" {
			output.Output = workspace
			return
		}

		info.Render(workspace, "", false)
	},
}
