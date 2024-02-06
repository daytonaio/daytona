// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_workspace

import (
	"context"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"os"

	"github.com/daytonaio/daytona/client"
	"github.com/daytonaio/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	ssh_tunnel_util "github.com/daytonaio/daytona/ssh_tunnel/util"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var SshProxyCmd = &cobra.Command{
	Use:    "ssh-proxy [PROFILE_ID] [WORKSPACE_NAME] [PROJECT_NAME]",
	Args:   cobra.RangeArgs(2, 3),
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		profileId := args[0]
		workspaceName := args[1]
		projectName := ""

		profile, err := c.GetProfile(profileId)
		if err != nil {
			log.Fatal(err)
		}

		conn, err := client.GetConn(&profile)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		if len(args) == 3 {
			projectName = args[2]
		} else {
			projectName, err = util.GetFirstWorkspaceProjectName(conn, workspaceName, projectName)
			if err != nil {
				log.Fatal(err)
			}
		}

		localSockFile := "/tmp/daytona/ssh_gateway.sock"
		errChan := make(chan error)
		tunnelStartedChan := make(chan bool, 1)

		if profile.Id != "default" {
			localSockFile = fmt.Sprintf("/tmp/daytona/daytona-ssh-%s-%s-%d.sock", workspaceName, projectName, rand.Intn(math.MaxInt32))

			tunnelStartedChan, errChan = ssh_tunnel_util.ForwardRemoteUnixSock(context.Background(), profile, localSockFile, "/tmp/daytona/ssh_gateway.sock")
		} else {
			tunnelStartedChan <- true
		}

		<-tunnelStartedChan

		socketConn, err := net.Dial("unix", localSockFile)
		if err != nil {
			log.Fatal(err)
		}

		//	pipe stdio to con
		go func() {
			_, err := io.Copy(os.Stdout, socketConn)
			if err != nil {
				errChan <- err
			}
			errChan <- nil
		}()

		go func() {
			_, err := io.Copy(socketConn, os.Stdin)
			if err != nil {
				errChan <- err
			}
			errChan <- nil
		}()

		if err := <-errChan; err != nil {
			log.Fatal(err)
		}
	},
}
