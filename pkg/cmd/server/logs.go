// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/views"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var followFlag bool
var fileFlag bool

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Output Daytona Server logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			return err
		}

		query := ""
		if followFlag {
			query += "follow=true"
		}

		switch {
		case fileFlag:
			return readServerLogFile()

		default:
			ws, res, err := apiclient.GetWebsocketConn(context.Background(), "/log/server", &activeProfile, &query)

			if err != nil {
				log.Error(apiclient.HandleErrorResponse(res, err))

				if activeProfile.Id != "default" {
					return nil
				}

				readLogsFile := true
				form := huh.NewForm(
					huh.NewGroup(
						huh.NewConfirm().Description("An error occurred while connecting to the server. Would you like to read from local log file instead?").
							Value(&readLogsFile),
					),
				).WithTheme(views.GetCustomTheme())
				formErr := form.Run()
				if formErr != nil {
					return formErr
				}

				if readLogsFile {
					return readServerLogFile()
				}
				return nil

			}

			for {
				_, msg, err := ws.ReadMessage()
				if err != nil {
					return nil
				}

				fmt.Println(string(msg))
			}
		}
	},
}

func readServerLogFile() error {
	views.RenderBorderedMessage("Reading from server log file...")
	cfg, err := server.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get server config: %w", err)
	}

	file, err := os.Open(cfg.LogFile.Path)
	if err != nil {
		return fmt.Errorf("while opening server logs: %w", err)
	}
	defer file.Close()
	msgChan := make(chan []byte)
	errChan := make(chan error)

	go util.ReadLog(context.Background(), file, followFlag, msgChan, errChan)

	for {
		select {
		case <-context.Background().Done():
			return nil
		case err := <-errChan:
			if err != nil {
				if err != io.EOF {
					return err
				}
				return nil
			}
		case msg := <-msgChan:
			fmt.Println(string(msg))
		}
	}
}

func init() {
	logsCmd.Flags().BoolVarP(&followFlag, "follow", "f", false, "Follow logs")
	logsCmd.Flags().BoolVar(&fileFlag, "file", false, "Read logs from local server log file")
}
