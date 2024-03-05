// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"fmt"
	"log"
	"strings"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
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

		output := strings.Join(util.ArrayMap(targets, func(t serverapiclient.ProviderTarget) string {
			return fmt.Sprintf("%s/%s: %s", *t.ProviderInfo.Name, *t.Name, *t.Options)
		}), "\n")

		view_util.RenderInfoMessage(output)
	},
}
