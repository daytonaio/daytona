// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/views/ide"
	"github.com/daytonaio/daytona/pkg/views/util"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var ideCmd = &cobra.Command{
	Use:   "ide",
	Short: "Choose the default IDE",
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

		c.DefaultIdeId = chosenIde.Id

		err = c.Save()
		if err != nil {
			log.Fatal(err)
		}

		util.RenderInfoMessage(fmt.Sprintf("Default IDE set to: %s", chosenIde.Name))
	},
}
