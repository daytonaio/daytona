// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"

	apikey "github.com/daytonaio/daytona/pkg/cmd/server/apikey"
	"github.com/daytonaio/daytona/pkg/cmd/server/daemon"
	. "github.com/daytonaio/daytona/pkg/cmd/server/provider"
	. "github.com/daytonaio/daytona/pkg/cmd/server/target"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/server/config"
	"github.com/daytonaio/daytona/pkg/server/frpc"
	"github.com/daytonaio/daytona/pkg/types"
	"github.com/daytonaio/daytona/pkg/views/util"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var runAsDaemon bool

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the server process",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if log.GetLevel() < log.InfoLevel {
			//	for now, force the log level to info when running the server
			log.SetLevel(log.InfoLevel)
		}

		if runAsDaemon {
			fmt.Println("Starting the Daytona Server daemon...")
			err := daemon.Start()
			if err != nil {
				log.Fatal(err)
			}
			c, err := config.GetConfig()
			if err != nil {
				log.Fatal(err)
			}
			printServerStartedMessage(c)
			return
		}

		errCh := make(chan error)

		err := server.Start(errCh)
		if err != nil {
			log.Fatal(err)
		}

		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			err := <-errCh
			if err != nil {
				log.Fatal(err)
			}
		}()

		if err := server.HealthCheck(); err != nil {
			log.Fatal(err)
		} else {
			printServerStartedMessage(c)
		}

		err = <-errCh
		if err != nil {
			log.Fatal(err)
		}
	},
}

func printServerStartedMessage(c *types.ServerConfig) {
	util.RenderBorderedMessage(fmt.Sprintf("Daytona Server running on port: %d.\nYou can now begin developing locally.\n\nIf you want to connect to the server remotely:\n\n1. Create an API key on this machine:\ndaytona server api-key new\n\n2. On the client machine run:\ndaytona profile add -a %s -k API_KEY", c.ApiPort, frpc.GetApiUrl(c)))
}

func init() {
	ServerCmd.PersistentFlags().BoolVarP(&runAsDaemon, "daemon", "d", false, "Run the server as a daemon")
	ServerCmd.AddCommand(configureCmd)
	ServerCmd.AddCommand(configCmd)
	ServerCmd.AddCommand(logsCmd)
	ServerCmd.AddCommand(TargetCmd)
	ServerCmd.AddCommand(ProviderCmd)
	ServerCmd.AddCommand(stopCmd)
	ServerCmd.AddCommand(restartCmd)
	ServerCmd.AddCommand(apikey.ApiKeyCmd)
}
