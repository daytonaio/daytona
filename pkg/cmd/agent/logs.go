// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/daytonaio/daytona/pkg/agent/config"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/spf13/cobra"
)

var followFlag bool

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Output Daytona Agent logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		logFilePath := config.GetLogFilePath()

		if logFilePath == nil {
			return errors.New("log file path not set")
		}

		file, err := os.Open(*logFilePath)
		if err != nil {
			return err
		}
		defer file.Close()

		msgChan := make(chan []byte)
		errChan := make(chan error)

		go logs.ReadLog(context.Background(), file, followFlag, msgChan, errChan)

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
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&followFlag, "follow", "f", false, "Follow logs")
}
