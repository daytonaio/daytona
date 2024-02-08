// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_ports

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/daytonaio/daytona/client"
	views_util "github.com/daytonaio/daytona/cmd/views/util"
	"github.com/daytonaio/daytona/config"
	"github.com/daytonaio/daytona/grpc/proto"
	"github.com/daytonaio/daytona/internal/util"
	ssh_tunnel_util "github.com/daytonaio/daytona/ssh_tunnel/util"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var portForwardCmd = &cobra.Command{
	Use:   "forward [WORKSPACE_NAME] [PROJECT_NAME] -p [PORT]",
	Short: "Forward port",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			log.Fatal(err)
		}

		conn, err := client.GetConn(nil)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		projectName := ""

		if len(args) == 2 {
			projectName = args[1]
		} else {
			projectName, err = util.GetFirstWorkspaceProjectName(conn, args[0], projectName)
			if err != nil {
				log.Fatal(err)
			}
		}

		hostPort, errChan := ForwardPort(conn, activeProfile, args[0], projectName, uint32(portArg))

		if hostPort == nil {
			if err = <-errChan; err != nil {
				log.Fatal(err)
			}
		} else {
			if *hostPort != uint32(portArg) {
				views_util.RenderInfoMessage(fmt.Sprintf("Port %d already in use.", portArg))
			}
			views_util.RenderInfoMessage(fmt.Sprintf("Port available at http://localhost:%d\n", *hostPort))

			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt)

			for {
				select {
				case <-c:
					stopPortForwardRequest := &proto.StopPortForwardRequest{
						WorkspaceName: args[0],
						Project:       projectName,
						Port:          uint32(portArg),
					}

					_, err = proto.NewPortsClient(conn).StopPortForward(context.Background(), stopPortForwardRequest)
					if err != nil {
						log.Fatal(err)
					}

					views_util.RenderInfoMessage("Port forwarding stopped")
					os.Exit(0)
				case err = <-errChan:
					if err != nil {
						log.Fatal(err)
					}
				}
			}
		}
	},
}

func init() {
	portForwardCmd.Flags().IntVarP(&portArg, "port", "p", 0, "Port to forward")
	portForwardCmd.MarkFlagRequired("port")
}

func ForwardPort(conn *grpc.ClientConn, activeProfile config.Profile, workspaceName string, projectName string, port uint32) (*uint32, chan error) {
	ctx := context.Background()

	client := proto.NewPortsClient(conn)

	forwardPortRequest := &proto.ForwardPortRequest{
		WorkspaceName: workspaceName,
		Project:       projectName,
		Port:          port,
	}

	errChan := make(chan error)

	response, err := client.ForwardPort(ctx, forwardPortRequest)
	if err != nil {
		go func() {
			errChan <- err
		}()
		return nil, errChan
	}

	if activeProfile.Name == "default" {
		go func() {
			errChan <- nil
		}()
		return &response.HostPort, errChan
	}

	hostPort, errChan := ssh_tunnel_util.ForwardRemoteTcpPort(activeProfile, uint16(response.HostPort))
	hostPort32 := uint32(hostPort)
	return &hostPort32, errChan
}
