// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"context"

	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views"
	gitprovider_view "github.com/daytonaio/daytona/pkg/views/gitprovider"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var GitProviderAddCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"new", "register", "update"},
	Short:   "Register a Git providers",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		apiClient, err := server.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		gitProviderData := serverapiclient.GitProvider{}
		gitProviderData.Id = new(string)
		gitProviderData.Username = new(string)
		gitProviderData.Token = new(string)
		gitProviderData.BaseApiUrl = new(string)

		gitprovider_view.GitProviderSelectionView(&gitProviderData, nil, false)

		if *gitProviderData.Id == "" {
			return
		}

		_, err = apiClient.GitProviderAPI.SetGitProvider(ctx).GitProviderConfig(gitProviderData).Execute()
		if err != nil {
			log.Fatal(err)
		}

		views.RenderInfoMessage("Git provider has been registered")
	},
}
