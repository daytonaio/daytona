// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"
	"fmt"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	cmd_common "github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	"github.com/daytonaio/daytona/pkg/cmd/targetconfig"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/views"
	logs_view "github.com/daytonaio/daytona/pkg/views/logs"
	target_view "github.com/daytonaio/daytona/pkg/views/target"
	targetconfig_view "github.com/daytonaio/daytona/pkg/views/targetconfig"
	"github.com/docker/docker/pkg/stringid"
	"github.com/spf13/cobra"
)

var targetCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a target",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			return err
		}

		createTargetDto, err := CreateTargetDtoFlow(ctx, TargetCreationParams{
			ApiClient:         apiClient,
			ActiveProfileName: activeProfile.Name,
		})
		if err != nil {
			if common.IsCtrlCAbort(err) {
				return nil
			} else {
				return err
			}
		}

		logsContext, stopLogs := context.WithCancel(context.Background())
		defer stopLogs()

		logs_view.SetupLongestPrefixLength([]string{createTargetDto.Name})

		logs_view.DisplayLogEntry(logs.LogEntry{
			TargetName: &createTargetDto.Name,
			Msg:        views.GetPrettyLogLine("Request submitted"),
		}, logs_view.STATIC_INDEX)

		go cmd_common.ReadTargetLogs(logsContext, cmd_common.ReadLogParams{
			Id:                    createTargetDto.Id,
			Label:                 &createTargetDto.Name,
			ServerUrl:             activeProfile.Api.Url,
			ApiKey:                activeProfile.Api.Key,
			Follow:                util.Pointer(true),
			SkipPrefixLengthSetup: true,
		})

		_, res, err := apiClient.TargetAPI.CreateTarget(ctx).Target(*createTargetDto).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		err = cmd_common.AwaitTargetState(createTargetDto.Id, apiclient.ResourceStateNameStarted)
		if err != nil {
			return err
		}

		// Ensure reading remaining logs is complete
		time.Sleep(100 * time.Millisecond)

		views.RenderInfoMessage(fmt.Sprintf("Target '%s' set successfully and will be used by default", createTargetDto.Name))
		return nil
	},
}

type TargetCreationParams struct {
	ApiClient         *apiclient.APIClient
	ActiveProfileName string
}

func CreateTargetDtoFlow(ctx context.Context, params TargetCreationParams) (*apiclient.CreateTargetDTO, error) {
	var targetConfigView *targetconfig_view.TargetConfigView

	targetConfigList, res, err := params.ApiClient.TargetConfigAPI.ListTargetConfigs(ctx).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	if len(targetConfigList) == 0 {
		targetConfigView, err = targetconfig.TargetConfigCreationFlow(ctx, params.ApiClient, params.ActiveProfileName)
		if err != nil {
			return nil, err
		}

		if targetConfigView == nil {
			return nil, common.ErrCtrlCAbort
		}
	} else {
		targetConfigView, err = targetconfig_view.GetTargetConfigFromPrompt(targetConfigList, params.ActiveProfileName, nil, true, "Use")
		if err != nil {
			return nil, err
		}

		if targetConfigView == nil {
			return nil, common.ErrCtrlCAbort
		}

		if targetConfigView.Name == targetconfig_view.NewTargetConfigName {
			targetConfigView, err = targetconfig.TargetConfigCreationFlow(ctx, params.ApiClient, params.ActiveProfileName)
			if err != nil {
				return nil, err
			}

			if targetConfigView == nil {
				return nil, common.ErrCtrlCAbort
			}
		}
	}

	if format.FormatFlag != "" {
		format.UnblockStdOut()
	}

	targetList, res, err := params.ApiClient.TargetAPI.ListTargets(ctx).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	var targetName string
	if targetConfigView != nil {
		targetName = targetConfigView.Name
	}

	target_view.SetTargetNameView(&targetName, util.ArrayMap(targetList, func(t apiclient.TargetDTO) string {
		return t.Name
	}))

	id := stringid.GenerateRandomID()
	id = stringid.TruncateID(id)

	return &apiclient.CreateTargetDTO{
		Id:               id,
		Name:             targetName,
		TargetConfigName: targetConfigView.Name,
	}, nil
}
