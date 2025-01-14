// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"context"
	"fmt"

	"github.com/charmbracelet/huh"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete a Git provider config",
	Aliases: common.GetAliases("delete"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		gitProviders, res, err := apiClient.GitProviderAPI.ListGitProviders(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		if len(gitProviders) == 0 {
			views_util.NotifyEmptyGitProviderList(true)
			return nil
		}

		if allFlag {
			if yesFlag {
				err = deleteAllGitProviders(gitProviders, apiClient)
				if err != nil {
					return err
				}
			} else {
				form := huh.NewForm(
					huh.NewGroup(
						huh.NewConfirm().
							Title("Delete all git providers?").
							Description("Are you sure you want to delete all git providers?").
							Value(&yesFlag),
					),
				).WithTheme(views.GetCustomTheme())
				err := form.Run()
				if err != nil {
					return err
				}

				if yesFlag {
					err = deleteAllGitProviders(gitProviders, apiClient)
					if err != nil {
						return err
					}
				} else {
					fmt.Println("Operation canceled.")
				}
			}

			return nil
		}

		selectedGitProvider := selection.GetGitProviderConfigFromPrompt(selection.GetGitProviderConfigParams{
			GitProviderConfigs: gitProviders,
			ActionVerb:         "Delete",
		})

		if selectedGitProvider == nil {
			return nil
		}

		selectedGitProviderText := fmt.Sprintf("%s (%s)", selectedGitProvider.ProviderId, selectedGitProvider.Alias)
		if !yesFlag {
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title(fmt.Sprintf("Delete git provider: %s?", selectedGitProviderText)).
						Description(fmt.Sprintf("Are you sure you want to delete the Git provider: %s?", selectedGitProviderText)).
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
			_, err = apiClient.GitProviderAPI.DeleteGitProvider(ctx, selectedGitProvider.Id).Execute()
			if err != nil {
				return err
			}

			views.RenderInfoMessage("Git provider has been deleted")
		}

		return nil
	},
}

var allFlag bool
var yesFlag bool

func init() {
	deleteCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Delete all Git providers")
	deleteCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Confirm deletion without prompt")
}

func deleteAllGitProviders(gitProviders []apiclient.GitProvider, apiClient *apiclient.APIClient) error {
	ctx := context.Background()
	for _, gitProvider := range gitProviders {
		_, err := apiClient.GitProviderAPI.DeleteGitProvider(ctx, gitProvider.Id).Execute()
		if err != nil {
			return err
		}
	}
	views.RenderInfoMessage("All Git providers have been removed")
	return nil
}
