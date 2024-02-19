// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_workspace

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/cli/connection"
	server_proto "github.com/daytonaio/daytona/common/grpc/proto"
	"github.com/golang/protobuf/ptypes/empty"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var CreateTempCmd = &cobra.Command{
	Use:   "create-temp",
	Short: "Create a workspace",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		conn, err := connection.Get(nil)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		client := server_proto.NewServerClient(conn)

		serverConfig, err := client.GetConfig(ctx, &empty.Empty{})
		if err != nil {
			log.Fatal(err)
		}

		if serverConfig.GitProviders == nil || len(serverConfig.GitProviders) == 0 {
			fmt.Println("No Git providers registered")
			return
		}

	},
}
