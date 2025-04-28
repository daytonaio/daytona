// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package image

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona-ai-saas/cli/apiclient"
	"github.com/daytonaio/daytona-ai-saas/cli/cmd/common"
	views_common "github.com/daytonaio/daytona-ai-saas/cli/views/common"
	views_util "github.com/daytonaio/daytona-ai-saas/cli/views/util"
	daytonaapiclient "github.com/daytonaio/daytona-ai-saas/daytonaapiclient"
	"github.com/spf13/cobra"
)

var BuildCmd = &cobra.Command{
	Use:   "build [IMAGE]",
	Short: "Build an image from a Dockerfile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if dockerfilePathFlag == "" {
			return fmt.Errorf("dockerfile path is required")
		}

		ctx := context.Background()
		imageName := args[0]

		err := common.ValidateImageName(imageName)
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
		_, res, err := apiClient.ImagesAPI.BuildImage(ctx).BuildImage(daytonaapiclient.BuildImage{
			Name:      imageName,
			BuildInfo: *createBuildInfoDto,
		}).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		err = views_util.WithInlineSpinner("Waiting for the image to build", func() error {
			return common.AwaitImageActive(ctx, apiClient, imageName)
		})
		if err != nil {
			return err
		}

		views_common.RenderInfoMessageBold(fmt.Sprintf("Use 'daytona sandbox create --image %s' to create a new sandbox using this image", imageName))
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
