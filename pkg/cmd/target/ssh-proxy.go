// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/cmd/tailscale"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/conversion"
	ssh_config "github.com/daytonaio/daytona/pkg/agent/ssh/config"
	"github.com/daytonaio/daytona/pkg/docker"
	"github.com/daytonaio/daytona/pkg/target/project"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var SshProxyCmd = &cobra.Command{
	Use:    "ssh-proxy [PROFILE_ID] [TARGET_ID] [PROJECT]",
	Args:   cobra.RangeArgs(2, 3),
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		profileId := args[0]
		targetId := args[1]
		projectName := ""

		profile, err := c.GetProfile(profileId)
		if err != nil {
			return err
		}

		if len(args) == 3 {
			projectName = args[2]
		} else {
			projectName, err = apiclient.GetFirstProjectName(targetId, projectName, &profile)
			if err != nil {
				return err
			}
		}

		target, err := apiclient.GetTarget(targetId, true)
		if err != nil {
			return err
		}

		if target.TargetConfig == "local" && profile.Id == "default" {
			// If the target is local, we directly access the ssh port through the container
			project := target.Projects[0]

			if project.Name != projectName {
				for _, p := range target.Projects {
					if p.Name == projectName {
						project = p
						break
					}
				}
			}

			cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
			if err != nil {
				return err
			}

			dockerClient := docker.NewDockerClient(docker.DockerClientConfig{
				ApiClient: cli,
			})

			containerName := dockerClient.GetProjectContainerName(conversion.ToProject(&project))

			ctx := context.Background()

			config := container.ExecOptions{
				AttachStdin:  true,
				AttachStderr: true,
				AttachStdout: true,
				Cmd:          []string{"daytona", "expose", fmt.Sprint(ssh_config.SSH_PORT)},
			}

			response, err := cli.ContainerExecCreate(ctx, containerName, config)
			if err != nil {
				return err
			}

			resp, err := cli.ContainerExecAttach(ctx, response.ID, container.ExecStartOptions{
				Tty: config.Tty,
			})

			if err != nil {
				return err
			}

			go func() {
				_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, resp.Reader)
				if err != nil {
					log.Fatal(err)
				}
			}()

			go func() {
				_, err := io.Copy(resp.Conn, os.Stdin)
				if err != nil {
					log.Fatal(err)
				}
			}()

			for {
				res, err := cli.ContainerExecInspect(ctx, response.ID)
				if err != nil {
					return err
				}

				if !res.Running {
					os.Exit(res.ExitCode)
				}

				time.Sleep(100 * time.Millisecond)
			}
		}

		tsConn, err := tailscale.GetConnection(&profile)
		if err != nil {
			return err
		}

		errChan := make(chan error)

		dialConn, err := tsConn.Dial(context.Background(), "tcp", fmt.Sprintf("%s:%d", project.GetProjectHostname(targetId, projectName), ssh_config.SSH_PORT))
		if err != nil {
			return err
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

		return <-errChan
	},
}
