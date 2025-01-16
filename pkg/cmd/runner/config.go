// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	view "github.com/daytonaio/daytona/pkg/views/runner"
	"github.com/spf13/cobra"

	"github.com/daytonaio/daytona/pkg/cmd/format"
	"github.com/daytonaio/daytona/pkg/runner"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Outputs Daytona Runner config",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := runner.GetConfig()
		if err != nil {
			return err
		}

		if format.FormatFlag != "" {
			formattedData := format.NewFormatter(config)
			formattedData.Print()
			return nil
		}

		view.RenderConfig(config, showKeyFlag)
		return nil
	},
}

var showKeyFlag bool

func init() {
	configCmd.Flags().BoolVarP(&showKeyFlag, "key", "k", false, "Show API Key")
	format.RegisterFormatFlag(configCmd)
}
