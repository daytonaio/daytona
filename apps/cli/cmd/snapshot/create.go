// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package snapshot

import (
	"context"
	"fmt"
	"strings"

	"github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/cmd/common"
	view_common "github.com/daytonaio/daytona/cli/views/common"
	views_util "github.com/daytonaio/daytona/cli/views/util"
	"github.com/daytonaio/daytona/daytonaapiclient"
	"github.com/spf13/cobra"
)

var CreateCmd = &cobra.Command{
	Use:     "create [SNAPSHOT]",
	Short:   "Create an snapshot",
	Args:    cobra.ExactArgs(1),
	Aliases: common.GetAliases("create"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		err := common.ValidateSnapshotName(args[0])
		if err != nil {
			return err
		}

		apiClient, err := apiclient.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		createSnapshotDto := daytonaapiclient.CreateSnapshot{
			Name: args[0],
		}

		if entrypointFlag != "" {
			createSnapshotDto.Entrypoint = strings.Split(entrypointFlag, " ")
		}

		snapshot, res, err := apiClient.SnapshotsAPI.CreateSnapshot(ctx).CreateSnapshot(createSnapshotDto).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		err = views_util.WithInlineSpinner("Waiting for the snapshot to be validated", func() error {
			return common.AwaitSnapshotState(ctx, apiClient, snapshot.Name, daytonaapiclient.SNAPSHOTSTATE_ACTIVE)
		})
		if err != nil {
			return err
		}

		view_common.RenderInfoMessageBold(fmt.Sprintf("Snapshot %s successfully created", snapshot.Name))

		view_common.RenderInfoMessage(fmt.Sprintf("%s Run 'daytona sandbox create --snapshot %s' to create a new sandbox using this snapshot", view_common.Checkmark, snapshot.Name))
		return nil
	},
}

var entrypointFlag string

func init() {
	CreateCmd.Flags().StringVarP(&entrypointFlag, "entrypoint", "e", "", "The entrypoint command for the snapshot")
}
