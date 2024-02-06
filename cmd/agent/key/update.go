// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_agent_key

import (
	"context"
	"dagent/client"
	"dagent/cmd/views/agent_update_key"
	"dagent/grpc/proto"
	"os"

	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var updateKeyCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the agent key",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := client.GetConn(nil)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		updateKeyView := agent_update_key.AgentUpdateKeyView{
			GenerateNewKey:   true,
			PathToPrivateKey: "",
		}

		if privateKeyPathFlag != "" {
			updateKeyView.GenerateNewKey = false
			updateKeyView.PathToPrivateKey = privateKeyPathFlag
		} else if generateFlag {
			updateKeyView.GenerateNewKey = true
		} else {
			agent_update_key.InteractiveForm(&updateKeyView)
		}

		agentClient := proto.NewAgentClient(conn)

		var response *proto.GetPublicKeyResponse

		if updateKeyView.GenerateNewKey {
			response, err = agentClient.GenerateKey(context.Background(), &proto.GenerateKeyRequest{})
			if err != nil {
				log.Fatal(err)
			}

		} else {
			privateKeyContent, err := os.ReadFile(updateKeyView.PathToPrivateKey)
			if err != nil {
				log.Fatal(err)
			}

			response, err = agentClient.SetKey(context.Background(), &proto.SetKeyRequest{
				PrivateKey: string(privateKeyContent),
			})
			if err != nil {
				log.Fatal(err)
			}
		}

		copyKeyAndNotify(response.PublicKey)
	},
}

var generateFlag bool
var privateKeyPathFlag string

func init() {
	updateKeyCmd.PersistentFlags().BoolVarP(&generateFlag, "generate", "g", false, "Generate a new key")
	updateKeyCmd.Flags().StringVarP(&privateKeyPathFlag, "private-key-path", "k", "", "Remote SSH private key path")
}
