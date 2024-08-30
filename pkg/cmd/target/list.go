// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"log"

	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	"github.com/daytonaio/daytona/pkg/views"
	list_view "github.com/daytonaio/daytona/pkg/views/target/list"
	"github.com/spf13/cobra"
)

var targetListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List targets",
	Args:    cobra.NoArgs,
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		targets, err := apiclient.GetTargetList()
		if err != nil {
			log.Fatal(err)
		}

		if len(targets) == 0 {
			views.RenderInfoMessageBold("No targets found")
			views.RenderInfoMessage("Use 'daytona target set' to add a target")
			return
		}

		if format.FormatFlag != "" {
			formattedData := format.NewFormatter(targets)
			formattedData.Print()
			return
		}

		list_view.ListTargets(targets)
	},
}

func init() {
	format.RegisterFormatFlag(targetListCmd)
}
