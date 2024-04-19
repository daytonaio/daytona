// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/cmd/output"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/views"
	view_util "github.com/daytonaio/daytona/pkg/views/util"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Output local Daytona Server config",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := server.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		apiUrl := util.GetFrpcApiUrl(config.Frps.Protocol, config.Id, config.Frps.Domain)
		output.Output = apiUrl

		output := ""
		output += "If you want to connect to the server remotely:\n\n"

		output += "1. Create an API key on this machine: "
		output += lipgloss.NewStyle().Foreground(views.Green).Render("daytona server api-key new") + "\n"
		output += "2. Add a profile on the client machine: \n\t"
		output += lipgloss.NewStyle().Foreground(views.Green).Render(fmt.Sprintf("daytona profile add -a %s -k API_KEY", apiUrl))
		view_util.RenderInfoMessage(output)
	},
}
