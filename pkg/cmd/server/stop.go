// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/daytonaio/daytona/pkg/cmd/common/daemon"
	"github.com/daytonaio/daytona/pkg/views"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops the Daytona Server daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		views.RenderInfoMessageBold("Stopping the Daytona Server daemon...")
		err := daemon.Stop(svcConfig)
		if errors.Is(err, daemon.ErrDaemonNotInstalled) {
			return fmt.Errorf("%w. First run 'daytona server' to start the server daemon", err)
		}

		return err
	},
}
