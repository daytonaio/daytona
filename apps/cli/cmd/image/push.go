// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package image

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/cmd/common"
	"github.com/daytonaio/daytona/cli/docker"
	views_common "github.com/daytonaio/daytona/cli/views/common"
	views_util "github.com/daytonaio/daytona/cli/views/util"
	"github.com/daytonaio/daytona/daytonaapiclient"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/spf13/cobra"
)

var PushCmd = &cobra.Command{
	Use:   "push [IMAGE]",
	Short: "Push local image",
	Long:  "Push a local image image to Daytona. To securely build it on our infastructure, use 'daytona image build'",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		sourceImage := args[0]

		err := common.ValidateImageName(sourceImage)
		if err != nil {
			return err
		}

		dockerClient, err := client.NewClientWithOpts(
			client.FromEnv,
			client.WithAPIVersionNegotiation(),
		)
		if err != nil {
			return fmt.Errorf("failed to create Docker client: %w", err)
		}
		defer dockerClient.Close()

		// Check if the image exists locally when not building
		if exists, err := docker.ImageExistsLocally(ctx, dockerClient, sourceImage); err != nil {
			return err
		} else if !exists {
			return fmt.Errorf("image '%s' not found locally. Please ensure the image exists and try again", sourceImage)
		}

		// Validate image architecture
		isArchAmd, err := docker.CheckAmdArchitecture(ctx, dockerClient, sourceImage)
		if err != nil {
			return fmt.Errorf("failed to check image architecture: %w", err)
		}
		if !isArchAmd {
			return fmt.Errorf("image '%s' is not compatible with AMD architecture", sourceImage)
		}

		apiClient, err := apiclient.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		tokenResponse, res, err := apiClient.DockerRegistryAPI.GetTransientPushAccess(ctx).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		encodedAuthConfig, err := json.Marshal(registry.AuthConfig{
			Username:      tokenResponse.Username,
			Password:      tokenResponse.Secret,
			ServerAddress: tokenResponse.RegistryUrl,
		})
		if err != nil {
			return fmt.Errorf("failed to marshal auth config: %w", err)
		}

		// Tag image
		targetImage := fmt.Sprintf("%s/%s/%s", tokenResponse.RegistryUrl, tokenResponse.Project, sourceImage)
		err = dockerClient.ImageTag(ctx, sourceImage, targetImage)
		if err != nil {
			return fmt.Errorf("failed to tag image: %w", err)
		}

		// Push image to transient registry
		pushReader, err := dockerClient.ImagePush(ctx, targetImage, image.PushOptions{
			RegistryAuth: base64.URLEncoding.EncodeToString(encodedAuthConfig),
		})
		if err != nil {
			return fmt.Errorf("failed to push image: %w", err)
		}
		defer pushReader.Close()

		err = jsonmessage.DisplayJSONMessagesStream(pushReader, os.Stdout, 0, true, nil)
		if err != nil {
			return err
		}

		createImage := daytonaapiclient.CreateImage{
			Name: targetImage,
		}

		// Poll until the image is really available on the registry
		// This is a workaround for harbor's delay in making newly created images available
		for {
			_, err := dockerClient.DistributionInspect(ctx, targetImage,
				base64.URLEncoding.EncodeToString(encodedAuthConfig))
			if err == nil {
				break
			}
			time.Sleep(time.Second)
		}

		_, res, err = apiClient.ImagesAPI.CreateImage(ctx).CreateImage(createImage).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		views_common.RenderInfoMessageBold(fmt.Sprintf("Successfully pushed %s to Daytona", sourceImage))

		err = views_util.WithInlineSpinner("Waiting for the image to be validated", func() error {
			return common.AwaitImageState(ctx, apiClient, targetImage, daytonaapiclient.IMAGESTATE_ACTIVE)
		})
		if err != nil {
			return err
		}

		views_common.RenderInfoMessage(fmt.Sprintf("%s  Use '%s' to create a new sandbox using this image", views_common.Checkmark, targetImage))
		return nil
	},
}
