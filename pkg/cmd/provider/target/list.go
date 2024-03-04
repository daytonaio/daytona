// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"strings"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views/provider"
	view_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var targetListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List provider targets",
	Args:    cobra.NoArgs,
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		pluginList, err := server.GetProviderList()
		if err != nil {
			log.Fatal(err)
		}

		selectedProvider := provider.GetProviderFromPrompt(pluginList, "Choose a Provider")

		if selectedProvider == nil {
			return
		}

		targets := strings.Join(util.ArrayMap(selectedProvider.Targets, func(t serverapiclient.TargetDTO) string {
			return *t.Name + ": " + *t.Options
		}), "\n")

		view_util.RenderInfoMessage(targets)
	},
}
