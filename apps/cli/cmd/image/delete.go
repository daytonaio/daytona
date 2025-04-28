// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package image

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona-ai-saas/cli/apiclient"
	"github.com/daytonaio/daytona-ai-saas/cli/cmd/common"
	view_common "github.com/daytonaio/daytona-ai-saas/cli/views/common"
	"github.com/spf13/cobra"
)

var DeleteCmd = &cobra.Command{
	Use:     "delete [IMAGE_ID]",
	Short:   "Delete an image",
	Args:    cobra.MaximumNArgs(1),
	Aliases: common.GetAliases("delete"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		var imageId string
		var imageName string

		apiClient, err := apiclient.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		imageList, res, err := apiClient.ImagesAPI.GetAllImages(ctx).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		if len(imageList.Items) == 0 {
			view_common.RenderInfoMessageBold("No images to delete")
			return nil
		}

		if len(args) == 0 {
			if allFlag {
				for _, image := range imageList.Items {
					res, err := apiClient.ImagesAPI.RemoveImage(ctx, image.Id).Execute()
					if err != nil {
						view_common.RenderInfoMessageBold(fmt.Sprintf("Failed to delete image %s: %s", image.Id, apiclient.HandleErrorResponse(res, err)))
					} else {
						view_common.RenderInfoMessageBold(fmt.Sprintf("Image %s deleted", image.Id))
					}
				}

				return nil
			}
			return cmd.Help()
		}

		for _, image := range imageList.Items {
			if image.Id == args[0] || image.Name == args[0] {
				imageId = image.Id
				imageName = image.Name
				break
			}
		}

		res, err = apiClient.ImagesAPI.RemoveImage(ctx, imageId).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		view_common.RenderInfoMessageBold(fmt.Sprintf("Image %s deleted", imageName))
		return nil
	},
}

var allFlag bool

func init() {
	DeleteCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Delete all images")
}
