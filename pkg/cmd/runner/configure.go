// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/runner"
	"github.com/daytonaio/daytona/pkg/views"
	runner_view "github.com/daytonaio/daytona/pkg/views/runner"
	"github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure Daytona Runner",
	RunE: func(cmd *cobra.Command, args []string) error {
		var configExisted bool

		config, err := runner.GetConfig()
		if err != nil {
			if errors.Is(err, runner.ErrConfigNotFound) {
				configExisted = false
				config, err = runner.GetDefaultConfig()
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		if idFlag != "" && nameFlag != "" && apiUrlFlag != "" && apiKeyFlag != "" {
			config.Id = idFlag
			config.Name = nameFlag
			config.ServerApiUrl = apiUrlFlag
			config.ServerApiKey = apiKeyFlag
		} else {
			config, err = runner_view.ConfigurationForm(config)
			if err != nil {
				return err
			}
		}

		config.ClientId = idFlag

		err = runner.Save(*config)
		if err != nil {
			return err
		}

		infoMessage := "Runner configuration updated. "

		if configExisted {
			infoMessage += "You need to restart the runner for the changes to take effect."
		} else {
			infoMessage += "To start running jobs, run 'daytona runner start'"
		}

		views.RenderContainerLayout(views.GetInfoMessage(infoMessage))
		return nil
	},
}

var idFlag string
var nameFlag string
var apiUrlFlag string
var apiKeyFlag string

func init() {
	configureCmd.Flags().StringVar(&idFlag, "id", "", "Runner ID")
	configureCmd.Flags().StringVar(&nameFlag, "name", "", "Runner Name")
	configureCmd.Flags().StringVar(&apiUrlFlag, "api-url", "", "Daytona Server API URL")
	configureCmd.Flags().StringVar(&apiKeyFlag, "api-key", "", "Runner API Key")
}
