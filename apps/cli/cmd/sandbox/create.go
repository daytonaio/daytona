// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/daytonaio/apiclient"
	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/cmd/common"
	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/cli/util"
	views_common "github.com/daytonaio/daytona/cli/views/common"
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

		apiClient, err := apiclient_cli.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		createSandbox := apiclient.NewCreateSandbox()

		// Add non-zero values to the request
		if snapshotFlag != "" {
			createSandbox.SetSnapshot(snapshotFlag)
		}
		if userFlag != "" {
			createSandbox.SetUser(userFlag)
		}
		if len(envFlag) > 0 {
			env := make(map[string]string)
			for _, e := range envFlag {
				parts := strings.SplitN(e, "=", 2)
				if len(parts) == 2 {
					env[parts[0]] = parts[1]
				}
			}
			createSandbox.SetEnv(env)
		}
		if len(labelsFlag) > 0 {
			labels := make(map[string]string)
			for _, l := range labelsFlag {
				parts := strings.SplitN(l, "=", 2)
				if len(parts) == 2 {
					labels[parts[0]] = parts[1]
				}
			}
			createSandbox.SetLabels(labels)
		}
		if publicFlag {
			createSandbox.SetPublic(true)
		}
		if classFlag != "" {
			createSandbox.SetClass(classFlag)
		}
		if targetFlag != "" {
			createSandbox.SetTarget(targetFlag)
		}
		if cpuFlag > 0 {
			createSandbox.SetCpu(cpuFlag)
		}
		if gpuFlag > 0 {
			createSandbox.SetGpu(gpuFlag)
		}
		if memoryFlag > 0 {
			createSandbox.SetMemory(memoryFlag)
		}
		if diskFlag > 0 {
			createSandbox.SetDisk(diskFlag)
		}
		if autoStopFlag >= 0 {
			createSandbox.SetAutoStopInterval(autoStopFlag)
		}
		if autoArchiveFlag >= 0 {
			createSandbox.SetAutoArchiveInterval(autoArchiveFlag)
		}
		createSandbox.SetAutoDeleteInterval(autoDeleteFlag)

		if dockerfileFlag != "" {
			createBuildInfoDto, err := common.GetCreateBuildInfoDto(ctx, dockerfileFlag, contextFlag)
			if err != nil {
				return err
			}
			createSandbox.SetBuildInfo(*createBuildInfoDto)
		}

		if len(volumesFlag) > 0 {
			volumes := make([]apiclient.SandboxVolume, 0, len(volumesFlag))
			for _, v := range volumesFlag {
				parts := strings.SplitN(v, ":", 2)
				if len(parts) == 2 {
					volumeId := parts[0]
					mountPath := parts[1]
					volume := apiclient.SandboxVolume{
						VolumeId:  volumeId,
						MountPath: mountPath,
					}
					volumes = append(volumes, volume)
				}
			}
			if len(volumes) > 0 {
				createSandbox.SetVolumes(volumes)
			}
		}

		var sandbox *apiclient.Sandbox

		sandbox, res, err := apiClient.SandboxAPI.CreateSandbox(ctx).CreateSandbox(*createSandbox).Execute()
		if err != nil {
			return apiclient_cli.HandleErrorResponse(res, err)
		}

		if sandbox.State != nil && *sandbox.State == apiclient.SANDBOXSTATE_PENDING_BUILD {
			c, err := config.GetConfig()
			if err != nil {
				return err
			}

			activeProfile, err := c.GetActiveProfile()
			if err != nil {
				return err
			}

			err = common.AwaitSandboxState(ctx, apiClient, sandbox.Id, apiclient.SANDBOXSTATE_BUILDING_SNAPSHOT)
			if err != nil {
				return err
			}

			logsContext, stopLogs := context.WithCancel(context.Background())
			defer stopLogs()

			go common.ReadBuildLogs(logsContext, common.ReadLogParams{
				Id:                   sandbox.Id,
				ServerUrl:            activeProfile.Api.Url,
				ServerApi:            activeProfile.Api,
				ActiveOrganizationId: activeProfile.ActiveOrganizationId,
				Follow:               util.Pointer(true),
				ResourceType:         common.ResourceTypeSandbox,
			})

			err = common.AwaitSandboxState(ctx, apiClient, sandbox.Id, apiclient.SANDBOXSTATE_STARTED)
			if err != nil {
				return err
			}

			// Wait for the last logs to be read
			time.Sleep(250 * time.Millisecond)
			stopLogs()
		}

		runnerDomain := sandbox.RunnerDomain
		if runnerDomain == nil {
			// Reload the sandbox info if the runner hadn't been assigned yet
			var getSandboxErr error
			sandbox, res, getSandboxErr = apiClient.SandboxAPI.GetSandbox(ctx, sandbox.Id).Execute()
			if getSandboxErr != nil {
				return apiclient_cli.HandleErrorResponse(res, getSandboxErr)
			}
			runnerDomain = sandbox.RunnerDomain
		}

		if runnerDomain == nil {
			return fmt.Errorf("failed to get runner domain")
		}

		previewUrl, res, err := apiClient.SandboxAPI.GetPortPreviewUrl(ctx, sandbox.Id, SANDBOX_TERMINAL_PORT).Execute()
		if err != nil {
			return apiclient_cli.HandleErrorResponse(res, err)
		}

		views_common.RenderInfoMessageBold(fmt.Sprintf("Sandbox is accessible at %s", views_common.LinkStyle.Render(previewUrl.Url)))
		return nil
	},
}

var (
	snapshotFlag    string
	userFlag        string
	envFlag         []string
	labelsFlag      []string
	publicFlag      bool
	classFlag       string
	targetFlag      string
	cpuFlag         int32
	gpuFlag         int32
	memoryFlag      int32
	diskFlag        int32
	autoStopFlag    int32
	autoArchiveFlag int32
	autoDeleteFlag  int32
	volumesFlag     []string
	dockerfileFlag  string
	contextFlag     []string
)

func init() {
	CreateCmd.Flags().StringVar(&snapshotFlag, "snapshot", "", "Snapshot to use for the sandbox")
	CreateCmd.Flags().StringVar(&userFlag, "user", "", "User associated with the sandbox")
	CreateCmd.Flags().StringArrayVarP(&envFlag, "env", "e", []string{}, "Environment variables (format: KEY=VALUE)")
	CreateCmd.Flags().StringArrayVarP(&labelsFlag, "label", "l", []string{}, "Labels (format: KEY=VALUE)")
	CreateCmd.Flags().BoolVar(&publicFlag, "public", false, "Make sandbox publicly accessible")
	CreateCmd.Flags().StringVar(&classFlag, "class", "", "Sandbox class type (small, medium, large)")
	CreateCmd.Flags().StringVar(&targetFlag, "target", "", "Target region (eu, us)")
	CreateCmd.Flags().Int32Var(&cpuFlag, "cpu", 0, "CPU cores allocated to the sandbox")
	CreateCmd.Flags().Int32Var(&gpuFlag, "gpu", 0, "GPU units allocated to the sandbox")
	CreateCmd.Flags().Int32Var(&memoryFlag, "memory", 0, "Memory allocated to the sandbox in MB")
	CreateCmd.Flags().Int32Var(&diskFlag, "disk", 0, "Disk space allocated to the sandbox in GB")
	CreateCmd.Flags().Int32Var(&autoStopFlag, "auto-stop", 15, "Auto-stop interval in minutes (0 means disabled)")
	CreateCmd.Flags().Int32Var(&autoArchiveFlag, "auto-archive", 10080, "Auto-archive interval in minutes (0 means the maximum interval will be used)")
	CreateCmd.Flags().Int32Var(&autoDeleteFlag, "auto-delete", -1, "Auto-delete interval in minutes (negative value means disabled, 0 means delete immediately upon stopping)")
	CreateCmd.Flags().StringArrayVarP(&volumesFlag, "volume", "v", []string{}, "Volumes to mount (format: VOLUME_NAME:MOUNT_PATH)")
	CreateCmd.Flags().StringVarP(&dockerfileFlag, "dockerfile", "f", "", "Path to Dockerfile for Sandbox snapshot")
	CreateCmd.Flags().StringArrayVarP(&contextFlag, "context", "c", []string{}, "Files or directories to include in the build context (can be specified multiple times)")

	CreateCmd.MarkFlagsMutuallyExclusive("snapshot", "dockerfile")
	CreateCmd.MarkFlagsMutuallyExclusive("snapshot", "context")
}
