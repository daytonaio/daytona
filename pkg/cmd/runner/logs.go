// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"context"
	"errors"
	"io"

	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/runner"
	logs_view "github.com/daytonaio/daytona/pkg/views/logs"
	"github.com/spf13/cobra"
)

var followFlag bool

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View runner logs",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := runner.GetConfig()
		if err != nil {
			return err
		}

		configDir, err := runner.GetConfigDir()
		if err != nil {
			return err
		}

		loggerFactory := logs.NewLoggerFactory(logs.LoggerFactoryConfig{
			LogsDir: runner.GetLogsDir(configDir),
		})

		logReader, err := loggerFactory.CreateLogReader(c.Id)
		if err != nil {
			return err
		}

		logs_view.SetupLongestPrefixLength([]string{c.Name})

		entryChan := make(chan interface{})
		errChan := make(chan error)
		go func() {
			logs.ReadJSONLog(context.Background(), logReader, followFlag, entryChan, errChan)
		}()

		go func() {
			for entry := range entryChan {
				logs_view.DisplayLogEntry(entry.(logs.LogEntry), logs_view.STATIC_INDEX)
			}
		}()

		err = <-errChan
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}

		return nil
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&followFlag, "follow", "f", false, "Follow logs")
}
