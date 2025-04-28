// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/daytonaio/daytona-ai-saas/cli/apiclient"
	"github.com/daytonaio/daytona-ai-saas/cli/cmd/common"
	views_common "github.com/daytonaio/daytona-ai-saas/cli/views/common"
	"github.com/daytonaio/daytona-ai-saas/cli/views/util"
	daytonaapiclient "github.com/daytonaio/daytona-ai-saas/daytonaapiclient"
	"github.com/spf13/cobra"
)

const SANDBOX_TERMINAL_PORT = 22222

var CreateCmd = &cobra.Command{
	Use:     "create [flags]",
	Short:   "Create a new sandbox",
	Args:    cobra.NoArgs,
	Aliases: common.GetAliases("create"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		createWorkspace := daytonaapiclient.NewCreateWorkspace()

		// Add non-zero values to the request
		if imageFlag != "" {
			createWorkspace.SetImage(imageFlag)
		}
		if userFlag != "" {
			createWorkspace.SetUser(userFlag)
		}
		if len(envFlag) > 0 {
			env := make(map[string]string)
			for _, e := range envFlag {
				parts := strings.SplitN(e, "=", 2)
				if len(parts) == 2 {
					env[parts[0]] = parts[1]
				}
			}
			createWorkspace.SetEnv(env)
		}
		if len(labelsFlag) > 0 {
			labels := make(map[string]string)
			for _, l := range labelsFlag {
				parts := strings.SplitN(l, "=", 2)
				if len(parts) == 2 {
					labels[parts[0]] = parts[1]
				}
			}
			createWorkspace.SetLabels(labels)
		}
		if publicFlag {
			createWorkspace.SetPublic(true)
		}
		if classFlag != "" {
			createWorkspace.SetClass(classFlag)
		}
		if targetFlag != "" {
			createWorkspace.SetTarget(targetFlag)
		}
		if cpuFlag > 0 {
			createWorkspace.SetCpu(cpuFlag)
		}
		if gpuFlag > 0 {
			createWorkspace.SetGpu(gpuFlag)
		}
		if memoryFlag > 0 {
			createWorkspace.SetMemory(memoryFlag)
		}
		if diskFlag > 0 {
			createWorkspace.SetDisk(diskFlag)
		}
		if autoStopFlag > 0 {
			createWorkspace.SetAutoStopInterval(autoStopFlag)
		}
		if dockerfileFlag != "" {
			createBuildInfoDto, err := common.GetCreateBuildInfoDto(ctx, dockerfileFlag, contextFlag)
			if err != nil {
				return err
			}
			createWorkspace.SetBuildInfo(*createBuildInfoDto)
		}

		if len(volumesFlag) > 0 {
			volumes := make([]daytonaapiclient.WorkspaceVolume, 0, len(volumesFlag))
			for _, v := range volumesFlag {
				parts := strings.SplitN(v, ":", 2)
				if len(parts) == 2 {
					volumeId := parts[0]
					mountPath := parts[1]
					volume := daytonaapiclient.WorkspaceVolume{
						VolumeId:  volumeId,
						MountPath: mountPath,
					}
					volumes = append(volumes, volume)
				}
			}
			if len(volumes) > 0 {
				createWorkspace.SetVolumes(volumes)
			}
		}

		var workspace *daytonaapiclient.Workspace

		err = util.WithInlineSpinner("Creating Sandbox", func() error {
			var res *http.Response
			var err error
			workspace, res, err = apiClient.WorkspaceAPI.CreateWorkspace(ctx).CreateWorkspace(*createWorkspace).Execute()
			if err != nil {
				return apiclient.HandleErrorResponse(res, err)
			}
			return nil
		})
		if err != nil {
			return err
		}

		var nodeDomain string
		if workspace.Info != nil && workspace.Info.ProviderMetadata != nil {
			metadata := make(map[string]interface{})
			if err := json.Unmarshal([]byte(*workspace.Info.ProviderMetadata), &metadata); err == nil {
				if domain, ok := metadata["nodeDomain"].(string); ok {
					nodeDomain = domain
				}
			}
		}

		sandboxUrl := fmt.Sprintf("https://%d-%s.%s", SANDBOX_TERMINAL_PORT, workspace.Id, nodeDomain)

		views_common.RenderInfoMessageBold(fmt.Sprintf("Sandbox is accessible at %s", views_common.LinkStyle.Render(sandboxUrl)))
		return nil
	},
}

var (
	imageFlag      string
	userFlag       string
	envFlag        []string
	labelsFlag     []string
	publicFlag     bool
	classFlag      string
	targetFlag     string
	cpuFlag        int32
	gpuFlag        int32
	memoryFlag     int32
	diskFlag       int32
	autoStopFlag   int32
	volumesFlag    []string
	dockerfileFlag string
	contextFlag    []string
)

func init() {
	CreateCmd.Flags().StringVar(&imageFlag, "image", "", "Image to use for the sandbox")
	CreateCmd.Flags().StringVar(&userFlag, "user", "", "User associated with the sandbox")
	CreateCmd.Flags().StringArrayVarP(&envFlag, "env", "e", []string{}, "Environment variables (format: KEY=VALUE)")
	CreateCmd.Flags().StringArrayVarP(&labelsFlag, "label", "l", []string{}, "Labels (format: KEY=VALUE)")
	CreateCmd.Flags().BoolVar(&publicFlag, "public", false, "Make sandbox publicly accessible")
	CreateCmd.Flags().StringVar(&classFlag, "class", "", "Workspace class type (small, medium, large)")
	CreateCmd.Flags().StringVar(&targetFlag, "target", "", "Target region (eu, us, asia)")
	CreateCmd.Flags().Int32Var(&cpuFlag, "cpu", 0, "CPU cores allocated to the sandbox")
	CreateCmd.Flags().Int32Var(&gpuFlag, "gpu", 0, "GPU units allocated to the sandbox")
	CreateCmd.Flags().Int32Var(&memoryFlag, "memory", 0, "Memory allocated to the sandbox in MB")
	CreateCmd.Flags().Int32Var(&diskFlag, "disk", 0, "Disk space allocated to the sandbox in GB")
	CreateCmd.Flags().Int32Var(&autoStopFlag, "auto-stop", 0, "Auto-stop interval in minutes (0 means disabled)")
	CreateCmd.Flags().StringArrayVarP(&volumesFlag, "volume", "v", []string{}, "Volumes to mount (format: VOLUME_NAME:MOUNT_PATH)")
	CreateCmd.Flags().StringVarP(&dockerfileFlag, "dockerfile", "f", "", "Path to Dockerfile for Sandbox image")
	CreateCmd.Flags().StringArrayVarP(&contextFlag, "context", "c", []string{}, "Files or directories to include in the build context (can be specified multiple times)")
}
