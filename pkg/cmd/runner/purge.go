// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"fmt"
	"net"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/pkg/runner"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/spf13/cobra"
)

var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Purges the runner",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := runner.GetConfig()
		if err != nil {
			return err
		}

		err = healthCheck(cfg.ApiPort)
		if err == nil {
			return runStopRunnerForm(cfg.ApiPort)
		}

		return purgeRunner()
	},
}

func healthCheck(apiPort int32) error {
	_, err := net.Dial("tcp", fmt.Sprintf(":%d", apiPort))

	return err
}

func runStopRunnerForm(apiPort int32) error {
	var runnerStoppedCheck bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Please stop the Daytona Runner before continuing by executing daytona runner stop command.").
				Description("Purging all data requires the Daytona Runner to be stopped.").
				Affirmative("Continue").
				Negative("Abort").
				Value(&runnerStoppedCheck),
		),
	).WithTheme(views.GetCustomTheme())

	err := form.Run()
	if err != nil {
		return err
	}

	if runnerStoppedCheck {
		err = healthCheck(apiPort)
		if err == nil {
			views.RenderInfoMessage("The Daytona Runner is still running. Please stop it before continuing.")
			return nil
		}
	} else {
		fmt.Println("Operation cancelled.")
		return nil
	}

	return purgeRunner()
}

func purgeRunner() error {
	err := runner.DeleteConfigDir()
	if err != nil {
		return err
	}

	views.RenderInfoMessageBold("The Daytona Runner has been purged from this device.")

	return nil
}
