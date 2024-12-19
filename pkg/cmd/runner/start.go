// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/common/daemon"
	"github.com/daytonaio/daytona/pkg/runner"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/kardianos/service"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var svcConfig = &service.Config{
	Name:        "DaytonaRunnerDaemon",
	DisplayName: "Daytona Runner",
	Description: "Daytona Runner daemon.",
	Arguments:   []string{"runner start-process"},
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the runner",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if log.GetLevel() < log.InfoLevel {
			//	for now, force the log level to info when running the server
			log.SetLevel(log.InfoLevel)
		}

		c, err := runner.GetConfig()
		if err != nil {
			return err
		}

		if c.ServerApiUrl == "" || c.ServerApiKey == "" {
			views.RenderInfoMessage("Configure the runner by using 'daytona runner configure' before starting it.")
			return nil
		}

		views.RenderInfoMessageBold("Starting the Daytona Runner daemon...")

		err = daemon.Start(c.LogFile.Path, svcConfig)
		if err != nil {
			return err
		}

		err = checkServerConnection(*c)
		if err != nil {
			return err
		}

		switch runtime.GOOS {
		case "linux":
			fmt.Printf("Use `loginctl enable-linger %s` to allow the service to run after logging out.\n", os.Getenv("USER"))
		}
		return nil
	},
}

func checkServerConnection(c runner.Config) error {
	apiClient, err := apiclient_util.GetRunnerApiClient(c.ServerApiUrl, c.ServerApiKey, c.ClientId, c.TelemetryEnabled)
	if err != nil {
		return err
	}

	for i := 0; i < 30; i++ {
		time.Sleep(1 * time.Second)
		_, _, err = apiClient.DefaultAPI.HealthCheck(context.Background()).Execute()
		if err != nil {
			continue
		}

		return nil
	}

	return err
}
