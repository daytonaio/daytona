// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_server_key

import (
	"context"
	"os"

	"github.com/daytonaio/daytona/cli/cmd/views/server_update_key"
	"github.com/daytonaio/daytona/cli/connection"
	"github.com/daytonaio/daytona/common/grpc/proto"

	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var updateKeyCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the server key",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := connection.Get(nil)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		updateKeyView := server_update_key.ServerUpdateKeyView{
			GenerateNewKey:   true,
			PathToPrivateKey: "",
		}

		if privateKeyPathFlag != "" {
			updateKeyView.GenerateNewKey = false
			updateKeyView.PathToPrivateKey = privateKeyPathFlag
		} else if generateFlag {
			updateKeyView.GenerateNewKey = true
		} else {
			server_update_key.InteractiveForm(&updateKeyView)
		}

		serverClient := proto.NewServerClient(conn)

		var response *proto.GetPublicKeyResponse

		if updateKeyView.GenerateNewKey {
			response, err = serverClient.GenerateKey(context.Background(), &proto.GenerateKeyRequest{})
			if err != nil {
				log.Fatal(err)
			}

		} else {
			privateKeyContent, err := os.ReadFile(updateKeyView.PathToPrivateKey)
			if err != nil {
				log.Fatal(err)
			}

			response, err = serverClient.SetKey(context.Background(), &proto.SetKeyRequest{
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
