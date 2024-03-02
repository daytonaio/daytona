// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ports

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var portForwardCmd = &cobra.Command{
	Use:   "forward [WORKSPACE_NAME] [PROJECT_NAME] -p [PORT]",
	Short: "Forward port",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		log.Fatal("Not implemented - no more need to go through server, use tailscale instead")
		// c, err := config.GetConfig()
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// activeProfile, err := c.GetActiveProfile()
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// conn, err := connection.GetGrpcConn(nil)
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// defer conn.Close()

		// projectName := ""

		// if len(args) == 2 {
		// 	projectName = args[1]
		// } else {
		// 	projectName, err = util.GetFirstWorkspaceProjectName(conn, args[0], projectName)
		// 	if err != nil {
		// 		log.Fatal(err)
		// 	}
		// }

		// hostPort, errChan := ForwardPort(conn, activeProfile, args[0], projectName, uint32(portArg))

		// if hostPort == nil {
		// 	if err = <-errChan; err != nil {
		// 		log.Fatal(err)
		// 	}
		// } else {
		// 	if *hostPort != uint32(portArg) {
		// 		views_util.RenderInfoMessage(fmt.Sprintf("Port %d already in use.", portArg))
		// 	}
		// 	views_util.RenderInfoMessage(fmt.Sprintf("Port available at http://localhost:%d\n", *hostPort))

		// 	c := make(chan os.Signal, 1)
		// 	signal.Notify(c, os.Interrupt)

		// 	for {
		// 		select {
		// 		case <-c:
		// 			stopPortForwardRequest := &proto.StopPortForwardRequest{
		// 				WorkspaceId: args[0],
		// 				Project:     projectName,
		// 				Port:        uint32(portArg),
		// 			}

		// 			_, err = proto.NewPortsClient(conn).StopPortForward(context.Background(), stopPortForwardRequest)
		// 			if err != nil {
		// 				log.Fatal(err)
		// 			}

		// 			views_util.RenderInfoMessage("Port forwarding stopped")
		// 			os.Exit(0)
		// 		case err = <-errChan:
		// 			if err != nil {
		// 				log.Fatal(err)
		// 			}
		// 		}
		// 	}
		// }
	},
}

func init() {
	portForwardCmd.Flags().IntVarP(&portArg, "port", "p", 0, "Port to forward")
	portForwardCmd.MarkFlagRequired("port")
}
