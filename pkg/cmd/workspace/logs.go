// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	logs_view "github.com/daytonaio/daytona/pkg/views/logs"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var followFlag bool
var workspaceFlag bool

var LogsCmd = &cobra.Command{
	Use:     "logs [WORKSPACE] [PROJECT]",
	Short:   "View logs for a workspace or project",
	GroupID: util.WORKSPACE_GROUP,
	Args:    cobra.RangeArgs(0, 2),
	Aliases: []string{"log"},
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			log.Fatal(err)
		}

		query := ""
		if followFlag {
			query += "follow=true"
		}

		ctx := context.Background()
		var workspaceId string
		var projectName string

		apiClient, err := apiclient_util.GetApiClient(&activeProfile)
		if err != nil {
			log.Fatal(err)
		}

		if len(args) == 0 {
			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
			if err != nil {
				log.Fatal(apiclient_util.HandleErrorResponse(res, err))
			}
			workspace := selection.GetWorkspaceFromPrompt(workspaceList, "Get Logs for")
			if workspace == nil {
				return
			}
			workspaceId = workspace.Id
		} else {
			workspace, err := apiclient_util.GetWorkspace(args[0])
			if err != nil {
				log.Fatal(err)
			}
			workspaceId = workspace.Id
		}

		if (len(args) == 0 || len(args) == 1) && !workspaceFlag {
			selectedProject, err := selectWorkspaceProject(workspaceId, &activeProfile)
			if err != nil {
				log.Fatal(err)
			}
			if selectedProject == nil {
				return
			}
			projectName = selectedProject.Name
		}

		if len(args) == 2 && !workspaceFlag {
			projectName = args[1]
		}

		if allFlag {
			handleWorkspaceLogs(activeProfile, workspaceId, projectName, query)
		} else {
			handleLogs(activeProfile, workspaceId, projectName, query)
		}
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) >= 1 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return getWorkspaceNameCompletions()
	},
}

func handleWorkspaceLogs(activeProfile config.Profile, workspaceId string, projectName string, query string) {
	workspaceLogsStarted := false
	stopLogs := false
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if !workspaceLogsStarted {
			time.Sleep(250 * time.Millisecond)
		}
		ws, res, err := apiclient_util.GetWebsocketConn(fmt.Sprintf("/log/workspace/%s/%s", workspaceId, projectName), &activeProfile, &query)
		if err != nil {
			log.Error(apiclient_util.HandleErrorResponse(res, err))
		}
		apiclient_util.ReadJSONLog(ws, &stopLogs, logs_view.PROJECT_INDEX)

	}()

	ws, res, err := apiclient_util.GetWebsocketConn(fmt.Sprintf("/log/workspace/%s", workspaceId), &activeProfile, &query)
	if err != nil {
		log.Error(apiclient_util.HandleErrorResponse(res, err))
	}
	apiclient_util.ReadJSONLog(ws, &stopLogs, logs_view.WORKSPACE_INDEX)
	workspaceLogsStarted = true
	wg.Wait()

}

func handleLogs(activeProfile config.Profile, workspaceId, projectName string, query string) {
	logEndpoint := fmt.Sprintf("/log/workspace/%s", workspaceId)
	logIndex := logs_view.WORKSPACE_INDEX
	stopLogs := false
	if projectName != "" {
		logEndpoint = fmt.Sprintf("%s/%s", logEndpoint, projectName)
		logIndex = logs_view.PROJECT_INDEX
	}

	ws, res, err := apiclient_util.GetWebsocketConn(logEndpoint, &activeProfile, &query)
	if err != nil {
		log.Error(apiclient_util.HandleErrorResponse(res, err))
	}

	apiclient_util.ReadJSONLog(ws, &stopLogs, logIndex)
}

func init() {
	LogsCmd.Flags().BoolVarP(&followFlag, "follow", "f", false, "Follow logs")
	LogsCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Show logs for workspace and project together")
	LogsCmd.Flags().BoolVarP(&workspaceFlag, "workspace", "w", false, "Show logs for workspace only")
}
