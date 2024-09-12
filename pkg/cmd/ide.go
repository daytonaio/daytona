// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	ide_util "github.com/daytonaio/daytona/pkg/ide"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/ide"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var ideCmd = &cobra.Command{
	Use:     "ide",
	Short:   "Choose the default IDE",
	GroupID: util.PROFILE_GROUP,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		ideList := config.GetIdeList()
		var chosenIde config.Ide

		chosenIdeId := ide.GetIdeIdFromPrompt(ideList)

		if chosenIdeId == "" {
			return
		}

		for _, ide := range ideList {
			if ide.Id == chosenIdeId {
				chosenIde = ide
			}
		}

		switch chosenIde.Id {
		case "vscode":
			ide_util.CheckAndAlertVSCodeInstalled()
		case "cursor":
			_, err := ide_util.GetCursorBinaryPath()
			if err != nil {
				log.Error(err)
			}
		case "fleet":
			if err := ide_util.CheckFleetInstallation(); err != nil {
				log.Error(err)
			}
		case "zed":
			_, err := ide_util.GetZedBinaryPath()
			if err != nil {
				log.Error(err)
			}
		}

		c.DefaultIdeId = chosenIde.Id

		telemetry.AdditionalData["ide"] = chosenIde.Id

		err = c.Save()
		if err != nil {
			log.Fatal(err)
		}

		content := fmt.Sprintf("%s %s", views.GetPropertyKey("Default IDE: "), chosenIde.Name)
		views.RenderContainerLayout(views.GetInfoMessage(content))
	},
}
