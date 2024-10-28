// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/jetbrains"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	target_util "github.com/daytonaio/daytona/pkg/cmd/target/util"
	"github.com/daytonaio/daytona/pkg/ide"
	"github.com/daytonaio/daytona/pkg/server/targets"
	"github.com/daytonaio/daytona/pkg/telemetry"
	ide_views "github.com/daytonaio/daytona/pkg/views/ide"
	"github.com/daytonaio/daytona/pkg/views/target/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var CodeCmd = &cobra.Command{
	Use:     "code [TARGET] [WORKSPACE]",
	Short:   "Open a target in your preferred IDE",
	Args:    cobra.RangeArgs(0, 2),
	Aliases: []string{"open"},
	GroupID: util.TARGET_GROUP,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		ctx := context.Background()
		var targetId string
		var workspaceName string
		var providerConfigId *string
		var ideId string
		var target *apiclient.TargetDTO

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
			targetList, res, err := apiClient.TargetAPI.ListTargets(ctx).Verbose(true).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(targetList) == 0 {
				views_util.NotifyEmptyTargetList(true)
				return nil
			}

			target = selection.GetTargetFromPrompt(targetList, "Open")
			if target == nil {
				return nil
			}
			targetId = target.Id
		} else {
			target, err = apiclient_util.GetTarget(url.PathEscape(args[0]), true)
			if err != nil {
				if strings.Contains(err.Error(), targets.ErrTargetNotFound.Error()) {
					log.Debug(err)
					return errors.New("target not found. You can see all target names by running the command `daytona list`")
				}
				return err
			}
			targetId = target.Id
		}

		if len(args) == 0 || len(args) == 1 {
			selectedWorkspace, err := selectTargetWorkspace(targetId, &activeProfile)
			if err != nil {
				return err
			}
			if selectedWorkspace == nil {
				return nil
			}

			workspaceName = selectedWorkspace.Name
			providerConfigId = selectedWorkspace.GitProviderConfigId
		}

		if len(args) == 2 {
			workspaceName = args[1]
			for _, workspace := range target.Workspaces {
				if workspace.Name == workspaceName {
					providerConfigId = workspace.GitProviderConfigId
					break
				}
			}
		}

		if ideFlag != "" {
			ideId = ideFlag
		}

		if !target_util.IsWorkspaceRunning(target, workspaceName) {
			wsRunningStatus, err := AutoStartTarget(target.Name, workspaceName)
			if err != nil {
				return err
			}
			if !wsRunningStatus {
				return nil
			}
		}

		providerMetadata := ""
		if ideId != "ssh" {
			providerMetadata, err = target_util.GetWorkspaceProviderMetadata(target, workspaceName)
			if err != nil {
				return err
			}
		}

		gpgKey, err := GetGitProviderGpgKey(apiClient, ctx, providerConfigId)
		if err != nil {
			log.Warn(err)
		}

		yesFlag, _ := cmd.Flags().GetBool("yes")
		ideList := config.GetIdeList()
		ide_views.RenderIdeOpeningMessage(target.Name, workspaceName, ideId, ideList)
		return openIDE(ideId, activeProfile, targetId, workspaceName, providerMetadata, yesFlag, gpgKey)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 1 {
			return getWorkspaceNameCompletions(cmd, args, toComplete)
		}

		return getTargetNameCompletions()
	},
}

func selectTargetWorkspace(targetId string, profile *config.Profile) (*apiclient.Workspace, error) {
	ctx := context.Background()

	apiClient, err := apiclient_util.GetApiClient(profile)
	if err != nil {
		return nil, err
	}

	targetInfo, res, err := apiClient.TargetAPI.GetTarget(ctx, targetId).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	if len(targetInfo.Workspaces) > 1 {
		selectedWorkspace := selection.GetWorkspaceFromPrompt(targetInfo.Workspaces, "Open")
		if selectedWorkspace == nil {
			return nil, nil
		}
		return selectedWorkspace, nil
	} else if len(targetInfo.Workspaces) == 1 {
		return &targetInfo.Workspaces[0], nil
	}

	return nil, errors.New("no workspaces found in target")
}

func openIDE(ideId string, activeProfile config.Profile, targetId string, workspaceName string, workspaceProviderMetadata string, yesFlag bool, gpgKey string) error {
	telemetry.AdditionalData["ide"] = ideId

	switch ideId {
	case "vscode":
		return ide.OpenVSCode(activeProfile, targetId, workspaceName, workspaceProviderMetadata, gpgKey)
	case "ssh":
		return ide.OpenTerminalSsh(activeProfile, targetId, workspaceName, gpgKey, nil)
	case "browser":
		return ide.OpenBrowserIDE(activeProfile, targetId, workspaceName, workspaceProviderMetadata, gpgKey)
	case "cursor":
		return ide.OpenCursor(activeProfile, targetId, workspaceName, workspaceProviderMetadata, gpgKey)
	case "jupyter":
		return ide.OpenJupyterIDE(activeProfile, targetId, workspaceName, workspaceProviderMetadata, yesFlag, gpgKey)
	case "fleet":
		return ide.OpenFleet(activeProfile, targetId, workspaceName, gpgKey)
	case "zed":
		return ide.OpenZed(activeProfile, targetId, workspaceName, gpgKey)
	default:
		_, ok := jetbrains.GetIdes()[jetbrains.Id(ideId)]
		if ok {
			return ide.OpenJetbrainsIDE(activeProfile, ideId, targetId, workspaceName, gpgKey)
		}
	}

	return errors.New("invalid IDE. Please choose one by running `daytona ide`")
}

var ideFlag string

func init() {
	ideList := config.GetIdeList()
	ids := make([]string, len(ideList))
	for i, ide := range ideList {
		ids[i] = ide.Id
	}
	ideListStr := strings.Join(ids, ", ")
	CodeCmd.Flags().StringVarP(&ideFlag, "ide", "i", "", fmt.Sprintf("Specify the IDE (%s)", ideListStr))

	CodeCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Automatically confirm any prompts")

}

func AutoStartTarget(targetId string, workspaceName string) (bool, error) {
	if !yesFlag {
		if !ide_views.RunStartTargetForm(targetId) {
			return false, nil
		}
	}

	apiClient, err := apiclient_util.GetApiClient(nil)
	if err != nil {
		return false, err
	}

	err = StartTarget(apiClient, targetId, workspaceName)
	if err != nil {
		return false, err
	}

	return true, nil
}
