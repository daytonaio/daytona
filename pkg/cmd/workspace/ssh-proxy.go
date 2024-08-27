// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/cmd/tailscale"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	ssh_config "github.com/daytonaio/daytona/pkg/agent/ssh/config"
	"github.com/daytonaio/daytona/pkg/workspace/project"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var SshProxyCmd = &cobra.Command{
	Use:    "ssh-proxy [PROFILE_ID] [WORKSPACE_ID] [PROJECT]",
	Args:   cobra.RangeArgs(2, 3),
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		profileId := args[0]
		workspaceId := args[1]
		projectName := ""

		profile, err := c.GetProfile(profileId)
		if err != nil {
			log.Fatal(err)
		}

		if len(args) == 3 {
			projectName = args[2]
		} else {
			projectName, err = apiclient.GetFirstWorkspaceProjectName(workspaceId, projectName, &profile)
			if err != nil {
				log.Fatal(err)
			}
		}

		tsConn, err := tailscale.GetConnection(&profile)
		if err != nil {
			log.Fatal(err)
		}

		errChan := make(chan error)

		dialConn, err := tsConn.Dial(context.Background(), "tcp", fmt.Sprintf("%s:%d", project.GetProjectHostname(workspaceId, projectName), ssh_config.SSH_PORT))
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
