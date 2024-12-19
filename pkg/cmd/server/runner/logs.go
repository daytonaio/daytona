// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"context"
	"errors"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	cmd_common "github.com/daytonaio/daytona/pkg/cmd/common"

	"github.com/spf13/cobra"
)

var followFlag bool

var logsCmd = &cobra.Command{
	Use:   "logs [RUNNER_ID]",
	Short: "View runner logs",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// apiClient, err := apiclient_util.GetApiClient(nil)
		// if err != nil {
		// 	return err
		// }

		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			return err
		}

		if len(args) == 0 {
			return errors.New("not implemented")
		}

		cmd_common.ReadRunnerLogs(ctx, cmd_common.ReadLogParams{
			Id:        args[0],
			Label:     &args[0],
			ServerUrl: activeProfile.Api.Url,
			ApiKey:    activeProfile.Api.Key,
			Index:     util.Pointer(0),
			Follow:    &followFlag,
		})
		return nil
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&followFlag, "follow", "f", false, "Follow logs")
}
