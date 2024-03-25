// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikey

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/daytonaio/daytona/pkg/server/db"
	"github.com/daytonaio/daytona/pkg/types"
	"github.com/daytonaio/daytona/pkg/views/server/apikey"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List API keys",
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		apiKeys, err := db.ListApiKeys()
		if err != nil {
			log.Fatal(err)
		}

		clientKeys := []*types.ApiKey{}
		for _, key := range apiKeys {
			if key.Type == types.ApiKeyTypeClient {
				clientKeys = append(clientKeys, key)
			}
		}

		apikey.ListApiKeys(clientKeys)
	},
}
