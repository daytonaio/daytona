// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_server_key

import (
	"context"

	views_util "github.com/daytonaio/daytona/cli/cmd/views/util"
	"github.com/daytonaio/daytona/cli/connection"
	"github.com/daytonaio/daytona/common/grpc/proto"
	"github.com/golang/protobuf/ptypes/empty"

	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var deleteKeyCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete the server private key",
	Args:    cobra.NoArgs,
	Aliases: []string{"remove", "rm"},
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := connection.GetGrpcConn(nil)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		serverClient := proto.NewServerClient(conn)

		_, err = serverClient.DeleteKey(context.Background(), &empty.Empty{})
		if err != nil {
			log.Fatal(err)
		}

		views_util.RenderInfoMessage("Key deleted")
	},
}
