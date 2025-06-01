// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package snapshot

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/cmd/common"
	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/cli/util"
	view_common "github.com/daytonaio/daytona/cli/views/common"
	views_util "github.com/daytonaio/daytona/cli/views/util"
	"github.com/daytonaio/daytona/daytonaapiclient"
	"github.com/spf13/cobra"
)

var CreateCmd = &cobra.Command{
	Use:     "create [SNAPSHOT]",
	Short:   "Create a snapshot",
	Args:    cobra.ExactArgs(1),
	Aliases: common.GetAliases("create"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		snapshotName := args[0]

		usingDockerfile := dockerfilePathFlag != ""
		usingImage := imageNameFlag != ""

		if !usingDockerfile && !usingImage {
			return fmt.Errorf("must specify either --dockerfile or --image")
		}

		apiClient, err := apiclient.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		createSnapshotDto := daytonaapiclient.CreateSnapshot{
			Name: snapshotName,
		}

		if usingDockerfile {
			createBuildInfoDto, err := common.GetCreateBuildInfoDto(ctx, dockerfilePathFlag, contextFlag)
			if err != nil {
				return err
			}
			createSnapshotDto.BuildInfo = createBuildInfoDto
		} else if usingImage {
			err := common.ValidateImageName(imageNameFlag)
			if err != nil {
				return err
			}
			createSnapshotDto.ImageName = &imageNameFlag
			if entrypointFlag != "" {
				createSnapshotDto.Entrypoint = strings.Split(entrypointFlag, " ")
			}
		} else if entrypointFlag != "" {
			createSnapshotDto.Entrypoint = strings.Split(entrypointFlag, " ")
		}

		// Send create request
		snapshot, res, err := apiClient.SnapshotsAPI.CreateSnapshot(ctx).CreateSnapshot(createSnapshotDto).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		// If we're building from a Dockerfile, show build logs
		if usingDockerfile {
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
		}

		err = views_util.WithInlineSpinner("Waiting for the snapshot to be validated", func() error {
			return common.AwaitSnapshotState(ctx, apiClient, snapshotName, daytonaapiclient.SNAPSHOTSTATE_ACTIVE)
		})
		if err != nil {
			return err
		}

		view_common.RenderInfoMessageBold(fmt.Sprintf("Snapshot %s successfully created", snapshotName))
		view_common.RenderInfoMessage(fmt.Sprintf("%s Run 'daytona sandbox create --snapshot %s' to create a new sandbox using this snapshot", view_common.Checkmark, snapshotName))
		return nil
	},
}

var (
	entrypointFlag     string
	imageNameFlag      string
	dockerfilePathFlag string
	contextFlag        []string
)

func init() {
	CreateCmd.Flags().StringVarP(&entrypointFlag, "entrypoint", "e", "", "The entrypoint command for the snapshot")
	CreateCmd.Flags().StringVarP(&imageNameFlag, "image", "i", "", "The image name for the snapshot")
	CreateCmd.Flags().StringVarP(&dockerfilePathFlag, "dockerfile", "f", "", "Path to Dockerfile to build")
	CreateCmd.Flags().StringArrayVarP(&contextFlag, "context", "c", []string{}, "Files or directories to include in the build context (can be specified multiple times)")

	CreateCmd.MarkFlagsMutuallyExclusive("image", "dockerfile")
	CreateCmd.MarkFlagsMutuallyExclusive("image", "context")
	CreateCmd.MarkFlagsMutuallyExclusive("entrypoint", "dockerfile")
	CreateCmd.MarkFlagsMutuallyExclusive("entrypoint", "context")
}
