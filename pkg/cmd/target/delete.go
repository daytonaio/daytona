// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/cmd/daytona/config"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/target/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var yesFlag bool
var forceFlag bool

var deleteCmd = &cobra.Command{
	Use:     "delete [TARGET]...",
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
		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			return err
		}

		if allFlag {
			return deleteAllTargetsView(ctx, activeProfile.Id, apiClient)
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
				target, _, err := apiclient_util.GetTarget(arg, false)
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

		var deleteTargetsFlag bool

		if !yesFlag {
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title(fmt.Sprintf("Delete target(s): [%s]?", strings.Join(targetDeleteListNames, ", "))).
						Description(fmt.Sprintf("Are you sure you want to delete the target(s): [%s]?", strings.Join(targetDeleteListNames, ", "))).
						Value(&deleteTargetsFlag),
				),
			).WithTheme(views.GetCustomTheme())

			err := form.Run()
			if err != nil {
				return err
			}
		}

		if !yesFlag && !deleteTargetsFlag {
			fmt.Println("Operation canceled.")
		} else {
			for _, target := range targetDeleteList {
				err := deleteTarget(ctx, activeProfile.Id, apiClient, target)
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
		return getAllTargetsByState(nil)
	},
}

func init() {
	deleteCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Delete all targets")
	deleteCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Confirm deletion without prompt")
	deleteCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Delete a target by force")
}

func deleteAllTargetsView(ctx context.Context, activeProfileId string, apiClient *apiclient.APIClient) error {
	var deleteAllTargetsFlag bool

	if yesFlag {
		fmt.Println("Deleting all targets.")
		err := deleteAllTargets(ctx, activeProfileId, apiClient)
		if err != nil {
			return err
		}
	} else {
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Delete all targets?").
					Description("Are you sure you want to delete all targets?").
					Value(&deleteAllTargetsFlag),
			),
		).WithTheme(views.GetCustomTheme())

		err := form.Run()
		if err != nil {
			return err
		}

		if deleteAllTargetsFlag {
			err := deleteAllTargets(ctx, activeProfileId, apiClient)
			if err != nil {
				return err
			}
		} else {
			fmt.Println("Operation canceled.")
		}
	}

	return nil
}

func deleteAllTargets(ctx context.Context, activeProfileId string, apiClient *apiclient.APIClient) error {
	targetList, res, err := apiClient.TargetAPI.ListTargets(ctx).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}

	for _, target := range targetList {
		err := deleteTarget(ctx, activeProfileId, apiClient, &target)
		if err != nil {
			log.Errorf("Failed to delete target %s: %v", target.Name, err)
			continue
		}
		views.RenderInfoMessage(fmt.Sprintf("- Target '%s' successfully deleted", target.Name))
	}
	return nil
}

func deleteTarget(ctx context.Context, activeProfileId string, apiClient *apiclient.APIClient, target *apiclient.TargetDTO) error {
	if len(target.Workspaces) > 0 {
		err := deleteWorkspacesForTarget(ctx, apiClient, target)
		if err != nil {
			return err
		}
	}

	message := fmt.Sprintf("Deleting target %s", target.Name)
	err := views_util.WithInlineSpinner(message, func() error {
		res, err := apiClient.TargetAPI.RemoveTarget(ctx, target.Id).Force(forceFlag).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		err = config.RemoveSshEntries(activeProfileId, target.Id)
		if err != nil {
			return err
		}

		err = common.AwaitTargetDeleted(target.Id)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func deleteWorkspacesForTarget(ctx context.Context, apiClient *apiclient.APIClient, target *apiclient.TargetDTO) error {
	var deleteWorkspacesFlag bool

	var targetWorkspacesNames []string
	for _, workspace := range target.Workspaces {
		targetWorkspacesNames = append(targetWorkspacesNames, workspace.Name)
	}

	if !yesFlag {
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(fmt.Sprintf("Target '%s' is used by %d workspace(s). Delete workspaces: [%s]?", target.Name, len(target.Workspaces), strings.Join(targetWorkspacesNames, ", "))).
					Description(fmt.Sprintf("Do you want to delete workspace(s): [%s]?", strings.Join(targetWorkspacesNames, ", "))).
					Value(&deleteWorkspacesFlag),
			),
		).WithTheme(views.GetCustomTheme())

		err := form.Run()
		if err != nil {
			return err
		}
	}

	if yesFlag || deleteWorkspacesFlag {
		for _, workspace := range target.Workspaces {
			err := common.DeleteWorkspace(ctx, apiClient, workspace.Id, workspace.Name, forceFlag)
			if err != nil {
				log.Errorf("Failed to delete workspace %s: %v", workspace.Name, err)
				continue
			}
			views.RenderInfoMessage(fmt.Sprintf("Workspace '%s' successfully deleted", workspace.Name))
		}
	}

	return nil
}
