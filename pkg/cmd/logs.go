// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	"github.com/spf13/cobra"
)

var (
    followFlag    bool
    workspaceFlag bool
    retryFlag     bool
    maxRetries    int
    fromTime      string
)

var logsCmd = &cobra.Command{
    Use:     "logs [WORKSPACE] [PROJECT_NAME]",
    Short:   "View logs for a workspace/project",
    Args:    cobra.RangeArgs(0, 2),
    GroupID: util.WORKSPACE_GROUP,
    Aliases: []string{"lg", "log"},
    Long: `Stream logs from a workspace or project with automatic reconnection support.
Examples:
  # Stream workspace logs
  daytona logs my-workspace

  # Stream project logs with auto-reconnect
  daytona logs my-workspace my-project --retry

  # Stream logs from a specific time
  daytona logs my-workspace --from="2024-12-18T22:00:00Z"

  # Follow logs with reconnection enabled
  daytona logs my-workspace --follow --retry`,
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

        var workspace *apiclient.WorkspaceDTO
        apiClient, err := apiclient_util.GetApiClient(&activeProfile)
        if err != nil {
            return err
        }

        var (
            showWorkspaceLogs = true
            projectNames      []string
        )

        // Parse fromTime if provided
        var fromTimePtr *time.Time
        if fromTime != "" {
            t, err := time.Parse(time.RFC3339, fromTime)
            if err != nil {
                return fmt.Errorf("invalid time format for --from flag: %w", err)
            }
            fromTimePtr = &t
        }

        // Handle workspace selection
        if len(args) == 0 {
            workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
            if err != nil {
                return apiclient_util.HandleErrorResponse(res, err)
            }
            if len(workspaceList) == 0 {
                views.RenderInfoMessage("The workspace list is empty. Start off by running 'daytona create'.")
                return nil
            }
            workspace = selection.GetWorkspaceFromPrompt(workspaceList, "Get Logs For")
        } else {
            workspace, err = apiclient_util.GetWorkspace(args[0], false)
            if err != nil {
                return err
            }
        }

        if workspace == nil {
            return errors.New("workspace not found")
        } else if len(workspace.Projects) == 0 {
            return errors.New("no projects found in workspace")
        }

        // Handle project selection
        if len(args) == 2 {
            projects := util.ArrayMap(workspace.Projects, func(p apiclient.Project) string {
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
                return errors.New("project not found in workspace")
            }
            projectNames = append(projectNames, args[1])
            if workspaceFlag {
                showWorkspaceLogs = true
            } else {
                showWorkspaceLogs = false
            }
        } else if !workspaceFlag {
            projectNames = util.ArrayMap(workspace.Projects, func(p apiclient.Project) string {
                return p.Name
            })
        }

        // Create log reader with retry support
        reader := apiclient_util.NewLogReader(&activeProfile, workspace.Id)
        if maxRetries > 0 {
            reader.SetMaxRetries(maxRetries)
        }

        // Show status message
        views.RenderInfoMessage("Connecting to logs...")

        // Set context based on follow flag
        if followFlag {
            ctx = context.Background()
        } else {
            ctx, _ = context.WithTimeout(ctx, 30*time.Second)
        }

        // Start reading logs with new retry mechanism
        err = reader.ReadWorkspaceLogs(ctx, projectNames, followFlag, showWorkspaceLogs, fromTimePtr)
        if err != nil {
            return fmt.Errorf("failed to read logs: %w", err)
        }

        return nil
    },
}

func init() {
    logsCmd.Flags().BoolVarP(&followFlag, "follow", "f", false, "Follow logs")
    logsCmd.Flags().BoolVarP(&workspaceFlag, "workspace", "w", false, "View workspace logs")
    logsCmd.Flags().BoolVar(&retryFlag, "retry", true, "Enable automatic reconnection")
    logsCmd.Flags().IntVar(&maxRetries, "max-retries", 5, "Maximum number of reconnection attempts")
    logsCmd.Flags().StringVar(&fromTime, "from", "", "Show logs from this time (RFC3339 format)")
}