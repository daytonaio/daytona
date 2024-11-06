// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/target/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var yesFlag bool
var forceFlag bool

var deleteCmd = &cobra.Command{
	Use:     "delete [TARGET]",
	Short:   "Delete a target",
	Aliases: []string{"remove", "rm"},
	RunE: func(cmd *cobra.Command, args []string) error {

		ctx := context.Background()

		var targetDeleteList = []*apiclient.TargetDTO{}
		var targetDeleteListNames = []string{}
		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		if allFlag {
			if yesFlag {
				fmt.Println("Deleting all targets.")
				err := DeleteAllTargets(workspaceList, forceFlag)
				if err != nil {
					return err
				}
			} else {
				form := huh.NewForm(
					huh.NewGroup(
						huh.NewConfirm().
							Title("Delete all targets?").
							Description("Are you sure you want to delete all targets?").
							Value(&yesFlag),
					),
				).WithTheme(views.GetCustomTheme())

				err := form.Run()
				if err != nil {
					return err
				}

				if yesFlag {
					err := DeleteAllTargets(workspaceList, forceFlag)
					if err != nil {
						return err
					}
				} else {
					fmt.Println("Operation canceled.")
				}
			}
			return nil
		}

		if len(args) == 0 {
			targetList, res, err := apiClient.TargetAPI.ListTargets(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(targetList) == 0 {
				views_util.NotifyEmptyTargetList(false)
				return nil
			}

			targetDeleteList = selection.GetTargetsFromPrompt(targetList, "Delete")
			for _, target := range targetDeleteList {
				targetDeleteListNames = append(targetDeleteListNames, target.Name)
			}
		} else {
			for _, arg := range args {
				target, err := apiclient_util.GetTarget(arg, false)
				if err != nil {
					log.Error(fmt.Sprintf("[ %s ] : %v", arg, err))
					continue
				}
				targetDeleteList = append(targetDeleteList, target)
				targetDeleteListNames = append(targetDeleteListNames, target.Name)
			}
		}

		if len(targetDeleteList) == 0 {
			return nil
		}

		if !yesFlag {
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title(fmt.Sprintf("Delete target(s): [%s]?", strings.Join(targetDeleteListNames, ", "))).
						Description(fmt.Sprintf("Are you sure you want to delete the target(s): [%s]?", strings.Join(targetDeleteListNames, ", "))).
						Value(&yesFlag),
				),
			).WithTheme(views.GetCustomTheme())

			err := form.Run()
			if err != nil {
				return err
			}
		}

		if !yesFlag {
			fmt.Println("Operation canceled.")
		} else {
			for _, target := range targetDeleteList {
				err := RemoveTarget(ctx, apiClient, target, workspaceList, forceFlag)
				if err != nil {
					log.Error(fmt.Sprintf("[ %s ] : %v", target.Name, err))
				} else {
					views.RenderInfoMessage(fmt.Sprintf("Target '%s' successfully deleted", target.Name))
				}
			}
		}
		return nil
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return getTargetNameCompletions()
	},
}

func init() {
	deleteCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Delete all targets")
	deleteCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Confirm deletion without prompt")
	deleteCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Delete a target by force")
}

func DeleteAllTargets(workspaceList []apiclient.WorkspaceDTO, force bool) error {
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
		err := RemoveTarget(ctx, apiClient, &target, workspaceList, force)
		if err != nil {
			log.Errorf("Failed to delete target %s: %v", target.Name, err)
			continue
		}
		views.RenderInfoMessage(fmt.Sprintf("- Target '%s' successfully deleted", target.Name))
	}
	return nil
}

func RemoveTarget(ctx context.Context, apiClient *apiclient.APIClient, target *apiclient.TargetDTO, workspaceList []apiclient.WorkspaceDTO, force bool) error {
	for _, workspace := range workspaceList {
		if workspace.TargetId == target.Id {
			return fmt.Errorf("target '%s' is in use by workspace '%s', please remove workspaces before deleting their target", target.Name, workspace.Name)
		}
	}

	message := fmt.Sprintf("Deleting target %s", target.Name)
	err := views_util.WithInlineSpinner(message, func() error {
		res, err := apiClient.TargetAPI.RemoveTarget(ctx, target.Id).Force(force).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
