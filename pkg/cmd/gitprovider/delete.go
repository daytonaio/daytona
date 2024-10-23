// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"context"
	"fmt"

	"github.com/charmbracelet/huh"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	"github.com/spf13/cobra"
)

var gitProviderDeleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"remove", "rm"},
	Short:   "Unregister a Git provider",
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
			views.RenderInfoMessage("No git providers registered")
			return nil
		}

		if allFlag {
			if yesFlag {
				err = removeAllGitProviders(gitProviders, apiClient)
				if err != nil {
					return err
				}
			} else {
				form := huh.NewForm(
					huh.NewGroup(
						huh.NewConfirm().
							Title("Remove all git providers?").
							Description("Are you sure you want to remove all git providers?").
							Value(&yesFlag),
					),
				).WithTheme(views.GetCustomTheme())
				err := form.Run()
				if err != nil {
					return err
				}

				if yesFlag {
					err = removeAllGitProviders(gitProviders, apiClient)
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
			ActionVerb:         "Remove",
		})

		if selectedGitProvider == nil {
			return nil
		}

		selectedGitProviderText := fmt.Sprintf("%s (%s)", selectedGitProvider.ProviderId, selectedGitProvider.Alias)
		if !yesFlag {
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title(fmt.Sprintf("Remove git provider: %s?", selectedGitProviderText)).
						Description(fmt.Sprintf("Are you sure you want to remove the git provider: %s?", selectedGitProviderText)).
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
			_, err = apiClient.GitProviderAPI.RemoveGitProvider(ctx, selectedGitProvider.Id).Execute()
			if err != nil {
				return err
			}

			views.RenderInfoMessage("Git provider has been removed")
		}

		return nil
	},
}

var allFlag bool
var yesFlag bool

func init() {
	gitProviderDeleteCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Remove all Git providers")
	gitProviderDeleteCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Confirm deletion without prompt")
}

func removeAllGitProviders(gitProviders []apiclient.GitProvider, apiClient *apiclient.APIClient) error {
	ctx := context.Background()
	for _, gitProvider := range gitProviders {
		_, err := apiClient.GitProviderAPI.RemoveGitProvider(ctx, gitProvider.Id).Execute()
		if err != nil {
			return err
		}
	}
	views.RenderInfoMessage("All Git providers have been removed")
	return nil
}
