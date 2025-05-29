// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package snapshot

import (
	"context"
	"fmt"
	"time"

	"github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/cmd/common"
	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/cli/util"
	views_common "github.com/daytonaio/daytona/cli/views/common"
	views_util "github.com/daytonaio/daytona/cli/views/util"
	daytonaapiclient "github.com/daytonaio/daytona/daytonaapiclient"
	"github.com/spf13/cobra"
)

var BuildCmd = &cobra.Command{
	Use:   "build [SNAPSHOT]",
	Short: "Build an snapshot from a Dockerfile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if dockerfilePathFlag == "" {
			return fmt.Errorf("dockerfile path is required")
		}

		ctx := context.Background()
		snapshotName := args[0]

		err := common.ValidateSnapshotName(snapshotName)
		if err != nil {
			return err
		}

		apiClient, err := apiclient.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		createBuildInfoDto, err := common.GetCreateBuildInfoDto(ctx, dockerfilePathFlag, contextFlag)
		if err != nil {
			return err
		}

		// Send build request
		snapshot, res, err := apiClient.SnapshotsAPI.BuildSnapshot(ctx).BuildSnapshot(daytonaapiclient.BuildSnapshot{
			Name:      snapshotName,
			BuildInfo: *createBuildInfoDto,
		}).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			return err
		}

		logsContext, stopLogs := context.WithCancel(context.Background())
		defer stopLogs()

		go common.ReadBuildLogs(logsContext, common.ReadLogParams{
			Id:                   snapshot.Id,
			ServerUrl:            activeProfile.Api.Url,
			ServerApi:            activeProfile.Api,
			ActiveOrganizationId: activeProfile.ActiveOrganizationId,
			Follow:               util.Pointer(true),
			ResourceType:         common.ResourceTypeSnapshot,
		})

		err = common.AwaitSnapshotState(ctx, apiClient, snapshotName, daytonaapiclient.SNAPSHOTSTATE_PENDING)
		if err != nil {
			return err
		}

		// Wait for the last logs to be read
		time.Sleep(250 * time.Millisecond)
		stopLogs()

		err = views_util.WithInlineSpinner("Waiting for the snapshot to be validated", func() error {
			return common.AwaitSnapshotState(ctx, apiClient, snapshotName, daytonaapiclient.SNAPSHOTSTATE_ACTIVE)
		})
		if err != nil {
			return err
		}

		views_common.RenderInfoMessageBold(fmt.Sprintf("Use 'daytona sandbox create --snapshot %s' to create a new sandbox using this snapshot", snapshotName))
		return nil
	},
}

var (
	dockerfilePathFlag string
	contextFlag        []string
)

func init() {
	BuildCmd.Flags().StringVarP(&dockerfilePathFlag, "dockerfile", "f", "", "Path to Dockerfile to build")
	BuildCmd.Flags().StringArrayVarP(&contextFlag, "context", "c", []string{}, "Files or directories to include in the build context (can be specified multiple times)")
}
