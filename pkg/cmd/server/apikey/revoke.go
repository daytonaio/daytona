// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikey

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/daytonaio/daytona/pkg/server/auth"
	"github.com/daytonaio/daytona/pkg/server/db"
	"github.com/daytonaio/daytona/pkg/views/server/apikey"
	"github.com/daytonaio/daytona/pkg/views/util"
)

var revokeCmd = &cobra.Command{
	Use:     "revoke [NAME]",
	Short:   "Revoke an API key",
	Aliases: []string{"r", "rm", "delete"},
	Args:    cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		var keyName string

		if len(args) == 1 {
			keyName = args[0]
		} else {
			apiKeys, err := db.ListApiKeys()
			if err != nil {
				log.Fatal(err)
			}

			keyName = apikey.GetApiKeyNameFromPrompt(apiKeys, "Select an API key to revoke")
			if keyName == "" {
				log.Fatal("No API key selected")
			}
		}

		err := auth.RevokeApiKey(keyName)
		if err != nil {
			log.Fatal(err)
		}

		util.RenderInfoMessage("API key revoked")
	},
}
