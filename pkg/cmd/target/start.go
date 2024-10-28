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
	target_util "github.com/daytonaio/daytona/pkg/cmd/target/util"
	"github.com/daytonaio/daytona/pkg/views"
	ide_views "github.com/daytonaio/daytona/pkg/views/ide"
	"github.com/daytonaio/daytona/pkg/views/target/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type TargetState string

const (
	TARGET_STATE_RUNNING TargetState = "Running"
	TARGET_STATE_STOPPED TargetState = "Unavailable"
)

var startWorkspaceFlag string
var allFlag bool
var codeFlag bool

var StartCmd = &cobra.Command{
	Use:     "start [TARGET]",
	Short:   "Start a target",
	Args:    cobra.RangeArgs(0, 1),
	GroupID: util.TARGET_GROUP,
	RunE: func(cmd *cobra.Command, args []string) error {
		var selectedTargetsNames []string
		var activeProfile config.Profile
		var ideId string
		var ideList []config.Ide
		var providerConfigId *string
		workspaceProviderMetadata := ""

		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if allFlag {
			return startAllTargets()
		}

		if len(args) == 0 {
			if startWorkspaceFlag != "" {
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
			selectedTargets := selection.GetTargetsFromPrompt(targetList, "Start")
			for _, targets := range selectedTargets {
				selectedTargetsNames = append(selectedTargetsNames, targets.Name)
			}
		} else {
			selectedTargetsNames = append(selectedTargetsNames, args[0])
		}

		if len(selectedTargetsNames) == 1 {
			targetName := selectedTargetsNames[0]
			var targetId string
			if codeFlag {
				c, err := config.GetConfig()
				if err != nil {
					return err
				}

				activeProfile, err = c.GetActiveProfile()
				if err != nil {
					return err
				}

				ideList = config.GetIdeList()
				ideId = c.DefaultIdeId

				targetInfo, res, err := apiClient.TargetAPI.GetTarget(ctx, targetName).Execute()
				if err != nil {
					return apiclient_util.HandleErrorResponse(res, err)
				}
				targetId = targetInfo.Id
				if startWorkspaceFlag == "" {
					startWorkspaceFlag = targetInfo.Workspaces[0].Name
					providerConfigId = targetInfo.Workspaces[0].GitProviderConfigId
				} else {
					for _, workspace := range targetInfo.Workspaces {
						if workspace.Name == startWorkspaceFlag {
							providerConfigId = workspace.GitProviderConfigId
							break
						}
					}
				}

				if ideId != "ssh" {
					workspaceProviderMetadata, err = target_util.GetWorkspaceProviderMetadata(targetInfo, targetInfo.Workspaces[0].Name)
					if err != nil {
						return err
					}
				}
			}

			err = StartTarget(apiClient, targetName, startWorkspaceFlag)
			if err != nil {
				return err
			}
			gpgKey, err := GetGitProviderGpgKey(apiClient, ctx, providerConfigId)
			if err != nil {
				log.Warn(err)
			}

			if startWorkspaceFlag == "" {
				views.RenderInfoMessage(fmt.Sprintf("Target '%s' started successfully", targetName))
			} else {
				views.RenderInfoMessage(fmt.Sprintf("Workspace '%s' from target '%s' started successfully", startWorkspaceFlag, targetName))

				if codeFlag {
					ide_views.RenderIdeOpeningMessage(targetName, startWorkspaceFlag, ideId, ideList)
					err = openIDE(ideId, activeProfile, targetId, startWorkspaceFlag, workspaceProviderMetadata, yesFlag, gpgKey)
					if err != nil {
						return err
					}
				}
			}
		} else {
			for _, target := range selectedTargetsNames {
				err := StartTarget(apiClient, target, "")
				if err != nil {
					log.Errorf("Failed to start target %s: %v\n\n", target, err)
					continue
				}
				views.RenderInfoMessage(fmt.Sprintf("- Target '%s' started successfully", target))
			}
		}
		return nil
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return getAllTargetsByState(TARGET_STATE_STOPPED)
	},
}

func init() {
	StartCmd.PersistentFlags().StringVarP(&startWorkspaceFlag, "workspace", "w", "", "Start a single workspace in the target (workspace name)")
	StartCmd.PersistentFlags().BoolVarP(&allFlag, "all", "a", false, "Start all targets")
	StartCmd.PersistentFlags().BoolVarP(&codeFlag, "code", "c", false, "Open the target in the IDE after target start")
	StartCmd.PersistentFlags().BoolVarP(&yesFlag, "yes", "y", false, "Automatically confirm any prompts")

	err := StartCmd.RegisterFlagCompletionFunc("workspace", getWorkspaceNameCompletions)
	if err != nil {
		log.Error("failed to register completion function: ", err)
	}
}

func startAllTargets() error {
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
		err := StartTarget(apiClient, target.Name, "")
		if err != nil {
			log.Errorf("Failed to start target %s: %v\n\n", target.Name, err)
			continue
		}

		views.RenderInfoMessage(fmt.Sprintf("- Target '%s' started successfully", target.Name))
	}
	return nil
}

func getWorkspaceNameCompletions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	ctx := context.Background()
	apiClient, err := apiclient_util.GetApiClient(nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}

	targetId := args[0]
	target, _, err := apiClient.TargetAPI.GetTarget(ctx, targetId).Execute()
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}

	var choices []string
	for _, workspace := range target.Workspaces {
		choices = append(choices, workspace.Name)
	}
	return choices, cobra.ShellCompDirectiveDefault
}

func getTargetNameCompletions() ([]string, cobra.ShellCompDirective) {
	ctx := context.Background()
	apiClient, err := apiclient_util.GetApiClient(nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	targetList, _, err := apiClient.TargetAPI.ListTargets(ctx).Execute()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var choices []string
	for _, v := range targetList {
		choices = append(choices, v.Name)
	}

	return choices, cobra.ShellCompDirectiveNoFileComp
}

func getAllTargetsByState(state TargetState) ([]string, cobra.ShellCompDirective) {
	ctx := context.Background()
	apiClient, err := apiclient_util.GetApiClient(nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	targetList, _, err := apiClient.TargetAPI.ListTargets(ctx).Execute()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var choices []string
	for _, target := range targetList {
		for _, workspace := range target.Workspaces {
			if workspace.State == nil {
				continue
			}
			if state == TARGET_STATE_RUNNING && workspace.State.Uptime != 0 {
				choices = append(choices, target.Name)
				break
			}
			if state == TARGET_STATE_STOPPED && workspace.State.Uptime == 0 {
				choices = append(choices, target.Name)
				break
			}
		}
	}

	return choices, cobra.ShellCompDirectiveNoFileComp
}

func StartTarget(apiClient *apiclient.APIClient, targetId, workspaceName string) error {
	ctx := context.Background()
	var workspaceNames []string
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

	target, err := apiclient_util.GetTarget(targetId, false)
	if err != nil {
		return err
	}
	if workspaceName != "" {
		workspaceNames = append(workspaceNames, workspaceName)
	} else {
		workspaceNames = util.ArrayMap(target.Workspaces, func(w apiclient.Workspace) string {
			return w.Name
		})
	}

	logsContext, stopLogs := context.WithCancel(context.Background())
	go apiclient_util.ReadTargetLogs(logsContext, activeProfile, target.Id, workspaceNames, true, true, &from)

	if workspaceName == "" {
		res, err := apiClient.TargetAPI.StartTarget(ctx, targetId).Execute()
		if err != nil {
			stopLogs()
			return apiclient_util.HandleErrorResponse(res, err)
		}
		time.Sleep(100 * time.Millisecond)
		stopLogs()
		return nil
	} else {
		res, err := apiClient.TargetAPI.StartWorkspace(ctx, targetId, workspaceName).Execute()
		if err != nil {
			stopLogs()
			return apiclient_util.HandleErrorResponse(res, err)
		}
		time.Sleep(100 * time.Millisecond)
		stopLogs()
		return nil
	}
}
