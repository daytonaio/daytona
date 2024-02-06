// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_agent_key

import (
	"context"
	"dagent/client"
	views_util "dagent/cmd/views/util"
	"dagent/grpc/proto"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/atotto/clipboard"
	log "github.com/sirupsen/logrus"
)

var KeyCmd = &cobra.Command{
	Use:   "key",
	Short: "Get the agent public key",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := client.GetConn(nil)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		agentClient := proto.NewAgentClient(conn)

		publicKey, err := agentClient.GetPublicKey(context.Background(), &proto.GetPublicKeyRequest{})
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
		fmt.Println("Agent Public Key")
		fmt.Println("Add this public key into your Git provider to enable cloning private repositories.")
	} else {
		views_util.RenderInfoMessageBold("Agent's Public Key has been copied to your clipboard.")
		views_util.RenderInfoMessage("Add it to your Git provider to enable cloning private repositories.")
	}

	fmt.Println()
	fmt.Println(key)
}
