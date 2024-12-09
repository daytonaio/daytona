// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"github.com/daytonaio/daytona/pkg/runner"
	"github.com/daytonaio/daytona/pkg/views"
	runner_view "github.com/daytonaio/daytona/pkg/views/runner"
	"github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure Daytona Runner",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := runner.GetConfig()
		if err != nil {
			return err
		}

		if idFlag != "" && apiUrlFlag != "" && apiKeyFlag != "" {
			config.Id = idFlag
			config.ServerApiUrl = apiUrlFlag
			config.ServerApiKey = apiKeyFlag
		} else {
			config, err = runner_view.ConfigurationForm(config)
			if err != nil {
				return err
			}
		}

		err = runner.Save(*config)
		if err != nil {
			return err
		}

		views.RenderContainerLayout(views.GetInfoMessage("Runner configuration updated. You need to restart the runner for the changes to take effect."))
		return nil
	},
}

var idFlag string
var apiUrlFlag string
var apiKeyFlag string

func init() {
	configureCmd.Flags().StringVar(&idFlag, "id", "", "Runner ID")
	configureCmd.Flags().StringVar(&apiUrlFlag, "api-url", "", "Daytona Server API URL")
	configureCmd.Flags().StringVar(&apiKeyFlag, "api-key", "", "Runner API Key")
}
