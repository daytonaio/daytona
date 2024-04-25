// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/agent/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var followFlag bool

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Output Daytona Agent logs",
	Run: func(cmd *cobra.Command, args []string) {
		logFilePath := config.GetLogFilePath()

		if logFilePath == nil {
			log.Fatal("Log file path not set")
		}

		file, err := os.Open(*logFilePath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		msgChan := make(chan []byte)
		errChan := make(chan error)

		go util.ReadLog(context.Background(), file, followFlag, msgChan, errChan)

		for {
			select {
			case <-context.Background().Done():
				return
			case err := <-errChan:
				if err != nil {
					if err != io.EOF {
						log.Fatal(err)
					}
					return
				}
			case msg := <-msgChan:
				fmt.Println(string(msg))
			}
		}
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&followFlag, "follow", "f", false, "Follow logs")
}
