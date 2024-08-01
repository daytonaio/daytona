// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"

	config_const "github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/common"
	gitprovider_view "github.com/daytonaio/daytona/pkg/views/gitprovider"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/create"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
)

type RepositoryWizardConfig struct {
	ApiClient           *apiclient.APIClient
	UserGitProviders    []apiclient.GitProvider
	Manual              bool
	MultiProject        bool
	SkipBranchSelection bool
	ProjectOrder        int
	SelectedRepos       map[string]int
}

func getRepositoryFromWizard(config RepositoryWizardConfig) (*apiclient.GitRepository, error) {
	var providerId string
	var namespaceId string
	var err error

	ctx := context.Background()

	if len(config.UserGitProviders) == 0 || config.Manual {
		return create.GetRepositoryFromUrlInput(config.MultiProject, config.ProjectOrder, config.ApiClient, config.SelectedRepos)
	}

	supportedProviders := config_const.GetSupportedGitProviders()
	var gitProviderViewList []gitprovider_view.GitProviderView

	for _, gitProvider := range config.UserGitProviders {
		for _, supportedProvider := range supportedProviders {
			if *gitProvider.Id == supportedProvider.Id {
				gitProviderViewList = append(gitProviderViewList,
					gitprovider_view.GitProviderView{
						Id:       *gitProvider.Id,
						Name:     supportedProvider.Name,
						Username: *gitProvider.Username,
					},
				)
			}
		}
	}
	providerId = selection.GetProviderIdFromPrompt(gitProviderViewList, config.ProjectOrder)
	if providerId == "" {
		return nil, common.ErrCtrlCAbort
	}

	if providerId == selection.CustomRepoIdentifier {
		return create.GetRepositoryFromUrlInput(config.MultiProject, config.ProjectOrder, config.ApiClient, config.SelectedRepos)
	}

	var namespaceList []apiclient.GitNamespace

	err = views_util.WithSpinner("Loading", func() error {
		namespaceList, _, err = config.ApiClient.GitProviderAPI.GetNamespaces(ctx, providerId).Execute()
		return err
	})
	if err != nil {
		return nil, err
	}

	if len(namespaceList) == 1 {
		namespaceId = *namespaceList[0].Id
	} else {
		namespaceId = selection.GetNamespaceIdFromPrompt(namespaceList, config.ProjectOrder)
		if namespaceId == "" {
			return nil, common.ErrCtrlCAbort
		}
	}

	var providerRepos []apiclient.GitRepository
	err = views_util.WithSpinner("Loading", func() error {
		providerRepos, _, err = config.ApiClient.GitProviderAPI.GetRepositories(ctx, providerId, namespaceId).Execute()
		return err
	})

	if err != nil {
		return nil, err
	}

	chosenRepo := selection.GetRepositoryFromPrompt(providerRepos, config.ProjectOrder, config.SelectedRepos)
	if chosenRepo == nil {
		return nil, common.ErrCtrlCAbort
	}

	if config.SkipBranchSelection {
		return chosenRepo, nil
	}

	return GetBranchFromWizard(BranchWizardConfig{
		ApiClient:    config.ApiClient,
		ProviderId:   providerId,
		NamespaceId:  namespaceId,
		ChosenRepo:   chosenRepo,
		ProjectOrder: config.ProjectOrder,
	})
}
