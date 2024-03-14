// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var followFlag bool

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Output Daytona Server logs",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			log.Fatal(err)
		}

		query := ""
		if followFlag {
			query = "?follow=true"
		}

		wsURL, err := apiclient.GetWebSocketUrl(activeProfile.Api.Url)
		if err != nil {
			log.Fatal(err)
		}

		wsURL = fmt.Sprintf("%s/log/server%s", wsURL, query)

		ws, res, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		for {
			_, msg, err := ws.ReadMessage()
			if err != nil {
				return
			}

			fmt.Println(string(msg))
		}
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&followFlag, "follow", "f", false, "Follow logs")
}
