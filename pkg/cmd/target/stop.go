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
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/target/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var stopProjectFlag string

var StopCmd = &cobra.Command{
	Use:     "stop [TARGET]",
	Short:   "Stop a target",
	GroupID: util.TARGET_GROUP,
	Args:    cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		timeFormat := time.Now().Format("2006-01-02 15:04:05")
		from, err := time.Parse("2006-01-02 15:04:05", timeFormat)
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

		if allFlag {
			return stopAllTargets(activeProfile, from)
		}

		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			if stopProjectFlag != "" {
				return cmd.Help()
			}
			targetList, res, err := apiClient.TargetAPI.ListTargets(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(targetList) == 0 {
				views_util.NotifyEmptyTargetList(true)
				return nil
			}

			selectedTargets := selection.GetTargetsFromPrompt(targetList, "Stop")

			for _, target := range selectedTargets {
				err := StopTarget(apiClient, target.Name, "")
				if err != nil {
					log.Errorf("Failed to stop target %s: %v\n\n", target.Name, err)
					continue
				}

				projectNames := util.ArrayMap(target.Projects, func(p apiclient.Project) string {
					return p.Name
				})
				apiclient_util.ReadTargetLogs(ctx, activeProfile, target.Id, projectNames, false, true, &from)
				views.RenderInfoMessage(fmt.Sprintf("- Target '%s' successfully stopped", target.Name))
			}
		} else {
			targetId := args[0]
			var projectNames []string

			err = StopTarget(apiClient, targetId, stopProjectFlag)
			if err != nil {
				return err
			}

			target, err := apiclient_util.GetTarget(targetId, false)
			if err != nil {
				return err
			}

			if startProjectFlag != "" {
				projectNames = append(projectNames, stopProjectFlag)
			} else {
				projectNames = util.ArrayMap(target.Projects, func(p apiclient.Project) string {
					return p.Name
				})
			}

			apiclient_util.ReadTargetLogs(ctx, activeProfile, target.Id, projectNames, false, true, &from)

			if stopProjectFlag != "" {
				views.RenderInfoMessage(fmt.Sprintf("Project '%s' from target '%s' successfully stopped", stopProjectFlag, targetId))
			} else {
				views.RenderInfoMessage(fmt.Sprintf("Target '%s' successfully stopped", targetId))
			}
		}
		return nil
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return getAllTargetsByState(TARGET_STATE_RUNNING)
	},
}

func init() {
	StopCmd.Flags().StringVarP(&stopProjectFlag, "project", "p", "", "Stop a single project in the target (project name)")
	StopCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Stop all targets")
}

func stopAllTargets(activeProfile config.Profile, from time.Time) error {
	ctx := context.Background()
	apiClient, err := apiclient_util.GetApiClient(nil)
	if err != nil {
		return err
	}

	targetList, res, err := apiClient.TargetAPI.ListTargets(ctx).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}

	for _, target := range targetList {
		err := StopTarget(apiClient, target.Name, "")
		if err != nil {
			log.Errorf("Failed to stop target %s: %v\n\n", target.Name, err)
			continue
		}

		projectNames := util.ArrayMap(target.Projects, func(p apiclient.Project) string {
			return p.Name
		})

		apiclient_util.ReadTargetLogs(ctx, activeProfile, target.Id, projectNames, false, true, &from)
		views.RenderInfoMessage(fmt.Sprintf("- Target '%s' successfully stopped", target.Name))
	}
	return nil
}

func StopTarget(apiClient *apiclient.APIClient, targetId, projectName string) error {
	ctx := context.Background()
	var message string
	var stopFunc func() error

	if projectName == "" {
		message = fmt.Sprintf("Target '%s' is stopping", targetId)
		stopFunc = func() error {
			res, err := apiClient.TargetAPI.StopTarget(ctx, targetId).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}
			return nil
		}
	} else {
		message = fmt.Sprintf("Project '%s' from target '%s' is stopping", projectName, targetId)
		stopFunc = func() error {
			res, err := apiClient.TargetAPI.StopProject(ctx, targetId, projectName).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}
			return nil
		}
	}

	err := views_util.WithInlineSpinner(message, stopFunc)
	if err != nil {
		return err
	}

	return nil
}
