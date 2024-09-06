// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspacemode

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"

	"github.com/daytonaio/daytona/internal/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var exposeCmd = &cobra.Command{
	Use:     "expose [PORT]",
	Short:   "Expose a local port over stdout - Used by the Daytona CLI to make direct connections to the project",
	Args:    cobra.ExactArgs(1),
	GroupID: util.WORKSPACE_GROUP,
	Run: func(cmd *cobra.Command, args []string) {
		port, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatal(err)
		}

		targetUrl := fmt.Sprintf("localhost:%d", port)

		dialConn, err := net.Dial("tcp", targetUrl)
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			_, err := io.Copy(os.Stdout, dialConn)
			if err != nil {
				log.Fatal(err)
			}
			dialConn.Close()
		}()

		go func() {
			_, err := io.Copy(dialConn, os.Stdin)
			if err != nil {
				log.Fatal(err)
			}
			dialConn.Close()
		}()

		select {}
	},
}
