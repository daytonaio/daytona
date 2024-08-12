// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"context"
	"fmt"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	gitprovider_view "github.com/daytonaio/daytona/pkg/views/gitprovider"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var GitProviderAddCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"new", "register", "update"},
	Short:   "Register a Git provider identity",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		gitProviderData := apiclient.GitProvider{}
		gitProviderData.Id = new(string)
		gitProviderData.Identity = new(string)
		gitProviderData.Username = new(string)
		gitProviderData.Token = new(string)
		gitProviderData.BaseApiUrl = new(string)

		gitprovider_view.GitProviderSelectionView(&gitProviderData, nil, false)

		if *gitProviderData.Id == "" || *gitProviderData.Identity == "" {
			return
		}

		res, err := apiClient.GitProviderAPI.SetGitProvider(ctx).GitProviderConfig(gitProviderData).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		views.RenderInfoMessage(fmt.Sprintf("Git provider identity '%s' has been registered", *gitProviderData.Identity))
	},
}
