// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/cmd/output"
	"github.com/daytonaio/daytona/pkg/server"
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

		fmt.Println(apiUrl)
	},
}
