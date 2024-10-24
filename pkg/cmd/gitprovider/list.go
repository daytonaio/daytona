// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"context"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	gitprovider_view "github.com/daytonaio/daytona/pkg/views/gitprovider"
	"github.com/daytonaio/daytona/pkg/views/gitprovider/list"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
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

		supportedProviders := config.GetSupportedGitProviders()
		gitProviderViewList := []gitprovider_view.GitProviderView{}

		for _, gitProvider := range gitProviders {
			for _, supportedProvider := range supportedProviders {
				if gitProvider.ProviderId == supportedProvider.Id {
					gitProviderView := gitprovider_view.GitProviderView{
						Id:         gitProvider.Id,
						ProviderId: gitProvider.ProviderId,
						Name:       supportedProvider.Name,
						Username:   gitProvider.Username,
						Alias:      gitProvider.Alias,
					}

					if gitProvider.BaseApiUrl != nil {
						gitProviderView.BaseApiUrl = *gitProvider.BaseApiUrl
					}

					if gitProvider.SigningMethod != nil {
						gitProviderView.SigningMethod = string(*gitProvider.SigningMethod)
					}

					gitProviderViewList = append(gitProviderViewList, gitProviderView)
				}
			}
		}

		if format.FormatFlag != "" {
			formattedData := format.NewFormatter(gitProviderViewList)
			formattedData.Print()
			return nil
		}

		if len(gitProviderViewList) == 0 {
			views_util.NotifyEmptyGitProviderList(true)
			return nil
		}

		list.ListGitProviders(gitProviderViewList)
		return nil
	},
}

func init() {
	format.RegisterFormatFlag(gitProviderListCmd)
}
