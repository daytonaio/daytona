// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	"github.com/daytonaio/daytona/pkg/views"
	gitprovider_view "github.com/daytonaio/daytona/pkg/views/gitprovider"
	"github.com/spf13/cobra"
)

var gitProviderListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "Lists your registered Git providers",
	RunE: func(cmd *cobra.Command, args []string) error {
		apiClient, err := apiclient.GetApiClient(nil)
		if err != nil {
			return err
		}

		gitProviders, res, err := apiClient.GitProviderAPI.ListGitProviders(context.Background()).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		if len(gitProviders) == 0 {
			views.RenderInfoMessage("No git providers registered. Add a new git provider by\npreparing a Personal Access Token and running 'daytona git-providers add'")
			return nil
		}

		views.RenderMainTitle("Registered Git Providers:")

		supportedProviders := config.GetSupportedGitProviders()
		var gitProviderViewList []gitprovider_view.GitProviderView

		for _, gitProvider := range gitProviders {
			for _, supportedProvider := range supportedProviders {
				if gitProvider.Id == supportedProvider.Id {
					signingMethod := ""
					if gitProvider.SigningMethod != nil {
						signingMethod = string(*gitProvider.SigningMethod)
					}

					gitProviderViewList = append(gitProviderViewList,
						gitprovider_view.GitProviderView{
							Id:            gitProvider.Id,
							Name:          supportedProvider.Name,
							Username:      gitProvider.Username,
							SigningMethod: signingMethod,
						},
					)
				}
			}
		}

		if format.FormatFlag != "" {
			formattedData := format.NewFormatter(gitProviderViewList)
			formattedData.Print()
			return nil
		}

		for _, gitProviderView := range gitProviderViewList {
			views.RenderListLine(fmt.Sprintf("%s (%s)", gitProviderView.Name, gitProviderView.Username))
		}
		return nil
	},
}

func init() {
	format.RegisterFormatFlag(gitProviderListCmd)
}
