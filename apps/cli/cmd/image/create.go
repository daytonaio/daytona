// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package image

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
	Use:     "create [IMAGE]",
	Short:   "Create an image",
	Args:    cobra.ExactArgs(1),
	Aliases: common.GetAliases("create"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		err := common.ValidateImageName(args[0])
		if err != nil {
			return err
		}

		apiClient, err := apiclient.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		createImageDto := daytonaapiclient.CreateImage{
			Name: args[0],
		}

		if entrypointFlag != "" {
			createImageDto.Entrypoint = strings.Split(entrypointFlag, " ")
		}

		image, res, err := apiClient.ImagesAPI.CreateImage(ctx).CreateImage(createImageDto).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		err = views_util.WithInlineSpinner("Waiting for the image to be validated", func() error {
			return common.AwaitImageActive(ctx, apiClient, image.Name)
		})
		if err != nil {
			return err
		}

		view_common.RenderInfoMessageBold(fmt.Sprintf("Image %s successfully created", image.Name))

		view_common.RenderInfoMessage(fmt.Sprintf("%s Run 'daytona sandbox create --image %s' to create a new sandbox using this image", view_common.Checkmark, image.Name))
		return nil
	},
}

var entrypointFlag string

func init() {
	CreateCmd.Flags().StringVarP(&entrypointFlag, "entrypoint", "e", "", "The entrypoint command for the image")
}
