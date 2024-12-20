// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/cmd/workspace/create"
	ide_views "github.com/daytonaio/daytona/pkg/views/ide"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var CodeCmd = &cobra.Command{
	Use:     "code [WORKSPACE]",
	Short:   "Open a workspace in your preferred IDE",
	Args:    cobra.RangeArgs(0, 2),
	Aliases: []string{"open"},
	GroupID: util.TARGET_GROUP,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		ctx := context.Background()
		var providerConfigId *string
		var ideId string
		var ws *apiclient.WorkspaceDTO

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			return err
		}

		ideId = c.DefaultIdeId

		apiClient, err := apiclient_util.GetApiClient(&activeProfile)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Verbose(true).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(workspaceList) == 0 {
				return errors.New("no workspaces found")
			}

			ws = selection.GetWorkspaceFromPrompt(workspaceList, "Open")
			if ws == nil {
				return nil
			}
		} else {
			var statusCode int
			ws, statusCode, err = apiclient_util.GetWorkspace(url.PathEscape(args[0]), true)
			if err != nil {
				if statusCode == http.StatusNotFound {
					log.Debug(err)
					return errors.New("workspace not found. You can see all workspace names by running the command `daytona list`")
				}
				return err
			}
		}
		if create.IdeFlag != "" {
			ideId = create.IdeFlag
		}

		if ws.State.Name == apiclient.ResourceStateNameStopped {
			wsRunningStatus, err := AutoStartWorkspace(*ws)
			if err != nil {
				return err
			}
			if !wsRunningStatus {
				return nil
			}
		}

		providerMetadata := *ws.Info.ProviderMetadata

		gpgKey, err := common.GetGitProviderGpgKey(apiClient, ctx, providerConfigId)
		if err != nil {
			log.Warn(err)
		}

		yesFlag, _ := cmd.Flags().GetBool("yes")
		ideList := config.GetIdeList()
		ide_views.RenderIdeOpeningMessage(ws.TargetId, ws.Name, ideId, ideList)
		return common.OpenIDE(ideId, activeProfile, ws.Id, providerMetadata, yesFlag, gpgKey)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return common.GetWorkspaceNameCompletions()
	},
}

func init() {
	ideList := config.GetIdeList()
	ids := make([]string, len(ideList))
	for i, ide := range ideList {
		ids[i] = ide.Id
	}
	ideListStr := strings.Join(ids, ", ")
	CodeCmd.Flags().StringVarP(&create.IdeFlag, "ide", "i", "", fmt.Sprintf("Specify the IDE (%s)", ideListStr))

	CodeCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Automatically confirm any prompts")

}

func AutoStartWorkspace(workspace apiclient.WorkspaceDTO) (bool, error) {
	if !yesFlag {
		if !ide_views.RunStartWorkspaceForm(workspace.Name) {
			return false, nil
		}
	}

	apiClient, err := apiclient_util.GetApiClient(nil)
	if err != nil {
		return false, err
	}

	err = StartWorkspace(apiClient, workspace)
	if err != nil {
		return false, err
	}

	return true, nil
}
