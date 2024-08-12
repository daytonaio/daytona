// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"log"

	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/output"
	"github.com/daytonaio/daytona/pkg/views"
	list_view "github.com/daytonaio/daytona/pkg/views/target/list"
	"github.com/spf13/cobra"
)

var formatFlag string
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

		if formatFlag != "" {
			display := output.NewOutputFormatter(targets, formatFlag)
			display.Print()
			return
		}

		list_view.ListTargets(targets)
	},
}

func init() {
	targetListCmd.PersistentFlags().StringVarP(&formatFlag, output.FormatFlagName, output.FormatFlagShortHand, formatFlag, output.FormatDescription)
	targetListCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if formatFlag != "" {
			output.BlockStdOut()
		}
	}
}
