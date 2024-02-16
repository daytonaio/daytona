// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_workspace

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/cli/connection"
	"github.com/daytonaio/daytona/internal/util"

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

		conn, err := connection.GetGrpcConn(&profile)
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

		tsConn, err := connection.GetTailscaleConn(&profile)
		if err != nil {
			log.Fatal(err)
		}

		errChan := make(chan error)

		dialConn, err := tsConn.Dial(context.Background(), "tcp", fmt.Sprintf("%s-%s:2222", workspaceName, projectName))
		if err != nil {
			log.Fatal(err)
		}

		//	pipe stdio to con
		go func() {
			_, err := io.Copy(os.Stdout, dialConn)
			if err != nil {
				errChan <- err
			}
			errChan <- nil
		}()

		go func() {
			_, err := io.Copy(dialConn, os.Stdin)
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
