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

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type TargetState string

const (
	TARGET_STATE_RUNNING TargetState = "Running"
	TARGET_STATE_STOPPED TargetState = "Unavailable"
)

var startProjectFlag string
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
		projectProviderMetadata := ""

		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if allFlag {
			return startAllTargets()
		}

		if len(args) == 0 {
			if startProjectFlag != "" {
				return cmd.Help()
			}
			targetList, res, err := apiClient.TargetAPI.ListTargets(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
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

				wsInfo, res, err := apiClient.TargetAPI.GetTarget(ctx, targetName).Execute()
				if err != nil {
					return apiclient_util.HandleErrorResponse(res, err)
				}
				targetId = wsInfo.Id
				if startProjectFlag == "" {
					startProjectFlag = wsInfo.Projects[0].Name
					providerConfigId = wsInfo.Projects[0].GitProviderConfigId
				} else {
					for _, project := range wsInfo.Projects {
						if project.Name == startProjectFlag {
							providerConfigId = project.GitProviderConfigId
							break
						}
					}
				}

				if ideId != "ssh" {
					projectProviderMetadata, err = target_util.GetProjectProviderMetadata(wsInfo, wsInfo.Projects[0].Name)
					if err != nil {
						return err
					}
				}
			}

			err = StartTarget(apiClient, targetName, startProjectFlag)
			if err != nil {
				return err
			}
			gpgKey, err := GetGitProviderGpgKey(apiClient, ctx, providerConfigId)
			if err != nil {
				log.Warn(err)
			}

			if startProjectFlag == "" {
				views.RenderInfoMessage(fmt.Sprintf("Target '%s' started successfully", targetName))
			} else {
				views.RenderInfoMessage(fmt.Sprintf("Project '%s' from target '%s' started successfully", startProjectFlag, targetName))

				if codeFlag {
					ide_views.RenderIdeOpeningMessage(targetName, startProjectFlag, ideId, ideList)
					err = openIDE(ideId, activeProfile, targetId, startProjectFlag, projectProviderMetadata, yesFlag, gpgKey)
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
	StartCmd.PersistentFlags().StringVarP(&startProjectFlag, "project", "p", "", "Start a single project in the target (project name)")
	StartCmd.PersistentFlags().BoolVarP(&allFlag, "all", "a", false, "Start all targets")
	StartCmd.PersistentFlags().BoolVarP(&codeFlag, "code", "c", false, "Open the target in the IDE after target start")
	StartCmd.PersistentFlags().BoolVarP(&yesFlag, "yes", "y", false, "Automatically confirm any prompts")

	err := StartCmd.RegisterFlagCompletionFunc("project", getProjectNameCompletions)
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

func getProjectNameCompletions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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
	for _, project := range target.Projects {
		choices = append(choices, project.Name)
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
		for _, project := range target.Projects {
			if project.State == nil {
				continue
			}
			if state == TARGET_STATE_RUNNING && project.State.Uptime != 0 {
				choices = append(choices, target.Name)
				break
			}
			if state == TARGET_STATE_STOPPED && project.State.Uptime == 0 {
				choices = append(choices, target.Name)
				break
			}
		}
	}

	return choices, cobra.ShellCompDirectiveNoFileComp
}

func StartTarget(apiClient *apiclient.APIClient, targetId, projectName string) error {
	ctx := context.Background()
	var projectNames []string
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
	if projectName != "" {
		projectNames = append(projectNames, projectName)
	} else {
		projectNames = util.ArrayMap(target.Projects, func(p apiclient.Project) string {
			return p.Name
		})
	}

	logsContext, stopLogs := context.WithCancel(context.Background())
	go apiclient_util.ReadTargetLogs(logsContext, activeProfile, target.Id, projectNames, true, true, &from)

	if projectName == "" {
		res, err := apiClient.TargetAPI.StartTarget(ctx, targetId).Execute()
		if err != nil {
			stopLogs()
			return apiclient_util.HandleErrorResponse(res, err)
		}
		time.Sleep(100 * time.Millisecond)
		stopLogs()
		return nil
	} else {
		res, err := apiClient.TargetAPI.StartProject(ctx, targetId, projectName).Execute()
		if err != nil {
			stopLogs()
			return apiclient_util.HandleErrorResponse(res, err)
		}
		time.Sleep(100 * time.Millisecond)
		stopLogs()
		return nil
	}
}
