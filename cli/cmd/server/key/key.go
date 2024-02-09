// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_server_key

import (
	"context"
	"fmt"

	views_util "github.com/daytonaio/daytona/cli/cmd/views/util"
	"github.com/daytonaio/daytona/cli/connection"
	"github.com/daytonaio/daytona/common/grpc/proto"
	"github.com/golang/protobuf/ptypes/empty"

	"github.com/spf13/cobra"

	"github.com/atotto/clipboard"
	log "github.com/sirupsen/logrus"
)

var KeyCmd = &cobra.Command{
	Use:   "key",
	Short: "Get the server public key",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := connection.Get(nil)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		serverClient := proto.NewServerClient(conn)

		publicKey, err := serverClient.GetPublicKey(context.Background(), &empty.Empty{})
		if err != nil {
			log.Fatal(err)
		}

		copyKeyAndNotify(publicKey.PublicKey)
	},
}

func init() {
	KeyCmd.AddCommand(updateKeyCmd)
	KeyCmd.AddCommand(deleteKeyCmd)
}

func copyKeyAndNotify(key string) {
	err := clipboard.WriteAll(key)
	if err != nil {
		views_util.RenderInfoMessage("Server Public Key")
		views_util.RenderInfoMessage("Add this public key into your Git provider to enable cloning private repositories.")
	} else {
		views_util.RenderInfoMessageBold("Server's Public Key has been copied to your clipboard.")
		views_util.RenderInfoMessage("Add it to your Git provider to enable cloning private repositories.")
	}

	fmt.Println()
	fmt.Println(key)
}
