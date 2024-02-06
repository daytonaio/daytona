// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_agent_key

import (
	"context"
	"dagent/client"
	"dagent/grpc/proto"
	"fmt"

	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var deleteKeyCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete the agent private key",
	Args:    cobra.NoArgs,
	Aliases: []string{"remove", "rm"},
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := client.GetConn(nil)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		agentClient := proto.NewAgentClient(conn)

		_, err = agentClient.DeleteKey(context.Background(), &proto.DeleteKeyRequest{})
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Key deleted")
	},
}
