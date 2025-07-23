// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package snapshot

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/daytonaio/apiclient"
	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/cmd/common"
	"github.com/daytonaio/daytona/cli/docker"
	views_common "github.com/daytonaio/daytona/cli/views/common"
	views_util "github.com/daytonaio/daytona/cli/views/util"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/spf13/cobra"
)

var PushCmd = &cobra.Command{
	Use:   "push [SNAPSHOT]",
	Short: "Push local snapshot",
	Long:  "Push a local Docker image to Daytona. To securely build it on our infrastructure, use 'daytona snapshot build'",
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

		apiClient, err := apiclient_cli.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		tokenResponse, res, err := apiClient.DockerRegistryAPI.GetTransientPushAccess(ctx).Execute()
		if err != nil {
			return apiclient_cli.HandleErrorResponse(res, err)
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

		createSnapshot := apiclient.NewCreateSnapshot(nameFlag)

		createSnapshot.SetImageName(targetImage)

		if entrypointFlag != "" {
			createSnapshot.SetEntrypoint(strings.Split(entrypointFlag, " "))
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

		if cpuFlag != 0 {
			createSnapshot.SetCpu(cpuFlag)
		}
		if memoryFlag != 0 {
			createSnapshot.SetMemory(memoryFlag)
		}
		if diskFlag != 0 {
			createSnapshot.SetDisk(diskFlag)
		}

		_, res, err = apiClient.SnapshotsAPI.CreateSnapshot(ctx).CreateSnapshot(*createSnapshot).Execute()
		if err != nil {
			return apiclient_cli.HandleErrorResponse(res, err)
		}

		views_common.RenderInfoMessageBold(fmt.Sprintf("Successfully pushed %s to Daytona", sourceImage))

		err = views_util.WithInlineSpinner("Waiting for the snapshot to be validated", func() error {
			return common.AwaitSnapshotState(ctx, apiClient, nameFlag, apiclient.SNAPSHOTSTATE_ACTIVE)
		})
		if err != nil {
			return err
		}

		views_common.RenderInfoMessage(fmt.Sprintf("%s  Use '%s' to create a new sandbox using this snapshot", views_common.Checkmark, nameFlag))
		return nil
	},
}

var (
	nameFlag string
)

func init() {
	PushCmd.Flags().StringVarP(&entrypointFlag, "entrypoint", "e", "", "The entrypoint command for the image")
	PushCmd.Flags().StringVarP(&nameFlag, "name", "n", "", "Specify the Snapshot name")
	PushCmd.Flags().Int32Var(&cpuFlag, "cpu", 0, "CPU cores that will be allocated to the underlying sandboxes (default: 1)")
	PushCmd.Flags().Int32Var(&memoryFlag, "memory", 0, "Memory that will be allocated to the underlying sandboxes in GB (default: 1)")
	PushCmd.Flags().Int32Var(&diskFlag, "disk", 0, "Disk space that will be allocated to the underlying sandboxes in GB (default: 3)")

	_ = PushCmd.MarkFlagRequired("name")
}
