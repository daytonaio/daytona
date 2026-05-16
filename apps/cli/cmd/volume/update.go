// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package volume

import (
        apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
        "github.com/daytonaio/daytona/cli/cmd/common"
        view_common "github.com/daytonaio/daytona/cli/views/common"
        "github.com/spf13/cobra"
)

var UpdateCmd = &cobra.Command{
        Use:     "update [ID]",
        Short:   "Update a volume",
        Args:    cobra.ExactArgs(1),
        Aliases: common.GetAliases("update"),
        RunE: func(cmd *cobra.Command, args []string) error {
                _, err := apiclient_cli.GetApiClient(nil, nil)
                if err != nil {
                        return err
                }

                // TODO: Once the volumes API client exposes a method for setting
                // the auto delete interval (matching POST /volumes/:volumeId/autodelete/:interval),
                // call it here using args[0] as the volume ID/name and autoDeleteIntervalFlag
                // as the interval in minutes.

                view_common.RenderInfoMessageBold("Volume " + args[0] + " auto-delete update is not yet wired to the API client")
                return nil
        },
}

func init() {
        UpdateCmd.Flags().Int32Var(&autoDeleteIntervalFlag, "auto-delete", -1, "Auto delete interval in minutes (-1 to disable, 0 on destroy)")
}