// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
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

var yesFlag bool
var forceFlag bool

var DeleteCmd = &cobra.Command{
	Use:     "delete [TARGET]",
	Short:   "Delete a target",
	GroupID: util.TARGET_GROUP,
	Aliases: []string{"remove", "rm"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if allFlag {
			if yesFlag {
				fmt.Println("Deleting all targets.")
				err := DeleteAllTargets(forceFlag)
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
					err := DeleteAllTargets(forceFlag)
					if err != nil {
						return err
					}
				} else {
					fmt.Println("Operation canceled.")
				}
			}
			return nil
		}

		ctx := context.Background()

		var targetDeleteList = []*apiclient.TargetDTO{}
		var targetDeleteListNames = []string{}
		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			targetList, res, err := apiClient.TargetAPI.ListTargets(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
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
				err := RemoveTarget(ctx, apiClient, target, forceFlag)
				if err != nil {
					log.Error(fmt.Sprintf("[ %s ] : %v", target.Name, err))
				}
				views.RenderInfoMessage(fmt.Sprintf("Target '%s' successfully deleted", target.Name))
			}
		}
		return nil
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return getTargetNameCompletions()
	},
}

func init() {
	DeleteCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Delete all targets")
	DeleteCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Confirm deletion without prompt")
	DeleteCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Delete a target by force")
}

func DeleteAllTargets(force bool) error {
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
		err := RemoveTarget(ctx, apiClient, &target, force)
		if err != nil {
			log.Errorf("Failed to delete target %s: %v", target.Name, err)
			continue
		}
		views.RenderInfoMessage(fmt.Sprintf("- Target '%s' successfully deleted", target.Name))
	}
	return nil
}

func RemoveTarget(ctx context.Context, apiClient *apiclient.APIClient, target *apiclient.TargetDTO, force bool) error {
	message := fmt.Sprintf("Deleting target %s", target.Name)
	err := views_util.WithInlineSpinner(message, func() error {
		res, err := apiClient.TargetAPI.RemoveTarget(ctx, target.Id).Force(force).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}
		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			return err
		}

		err = config.RemoveTargetSshEntries(activeProfile.Id, target.Id)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
