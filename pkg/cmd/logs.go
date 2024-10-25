// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"errors"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/target/selection"
	"github.com/spf13/cobra"
)

var followFlag bool
var targetFlag bool

var logsCmd = &cobra.Command{
	Use:     "logs [TARGET] [PROJECT_NAME]",
	Short:   "View logs for a target/project",
	Args:    cobra.RangeArgs(0, 2),
	GroupID: util.TARGET_GROUP,
	Aliases: []string{"lg", "log"},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			return err
		}

		var target *apiclient.TargetDTO
		apiClient, err := apiclient_util.GetApiClient(&activeProfile)
		if err != nil {
			return err
		}

		var (
			showTargetLogs = true
			projectNames   []string
		)

		if len(args) == 0 {
			targetList, res, err := apiClient.TargetAPI.ListTargets(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}
			if len(targetList) == 0 {
				views.RenderInfoMessage("The target list is empty. Start off by running 'daytona create'.")
				return nil
			}
			target = selection.GetTargetFromPrompt(targetList, "Get Logs For")
		} else {
			target, err = apiclient_util.GetTarget(args[0], false)
			if err != nil {
				return err
			}
		}

		if target == nil {
			return errors.New("target not found")
		} else if len(target.Projects) == 0 {
			return errors.New("no projects found in target")
		}

		if len(args) == 2 {
			projects := util.ArrayMap(target.Projects, func(p apiclient.Project) string {
				return p.Name
			})
			var found bool
			for _, project := range projects {
				if project == args[1] {
					found = true
					break
				}
			}
			if !found {
				return errors.New("project not found in target")
			}
			projectNames = append(projectNames, args[1])
			if targetFlag {
				showTargetLogs = true
			} else {
				showTargetLogs = false
			}
		} else if !targetFlag {
			projectNames = util.ArrayMap(target.Projects, func(p apiclient.Project) string {
				return p.Name
			})
		}

		apiclient_util.ReadTargetLogs(ctx, activeProfile, target.Id, projectNames, followFlag, showTargetLogs, nil)

		return nil
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&followFlag, "follow", "f", false, "Follow logs")
	logsCmd.Flags().BoolVarP(&targetFlag, "target", "w", false, "View target logs")
}
