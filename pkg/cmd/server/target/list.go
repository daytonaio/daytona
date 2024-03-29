// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"log"

	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/cmd/output"
	list_view "github.com/daytonaio/daytona/pkg/views/server/target/list"
	view_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/spf13/cobra"
)

var targetListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List targets",
	Args:    cobra.NoArgs,
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		targets, err := server.GetTargetList()
		if err != nil {
			log.Fatal(err)
		}

		if len(targets) == 0 {
			view_util.RenderInfoMessageBold("No targets found")
			view_util.RenderInfoMessage("Use 'daytona target set' to add a target")
			return
		}

		if output.FormatFlag != "" {
			output.Output = targets
			return
		}

		list_view.ListTargets(targets)
	},
}
