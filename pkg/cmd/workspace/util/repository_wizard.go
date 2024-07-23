// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"errors"
	"net/url"

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
	MultiProject        bool
	SkipBranchSelection bool
	ProjectOrder        int
	SelectedRepos       map[string]int
}

type BranchWizardConfig struct {
	ApiClient    *apiclient.APIClient
	ProviderId   string
	NamespaceId  string
	ChosenRepo   *apiclient.GitRepository
	ProjectOrder int
}

func getRepositoryFromWizard(config RepositoryWizardConfig) (*apiclient.GitRepository, error) {
	var providerId string
	var namespaceId string
	var err error

	ctx := context.Background()

	if len(config.UserGitProviders) == 0 {
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

func GetBranchFromWizardFromRepo(branchList []apiclient.GitBranch) (*apiclient.GitBranch, error) {
	branch := selection.GetBranchFromPrompt(branchList, 0)
	if branch == nil {
		return nil, errors.New("must select a branch")
	}

	return branch, nil
}

func GetBranchFromWizard(config BranchWizardConfig) (*apiclient.GitRepository, error) {
	var branchList []apiclient.GitBranch
	var checkoutOptions []selection.CheckoutOption
	var err error
	ctx := context.Background()

	err = views_util.WithSpinner("Loading", func() error {
		branchList, _, err = config.ApiClient.GitProviderAPI.GetRepoBranches(ctx, config.ProviderId, config.NamespaceId, url.QueryEscape(*config.ChosenRepo.Id)).Execute()
		return err
	})

	if err != nil {
		return nil, err
	}

	if len(branchList) == 0 {
		return nil, errors.New("no branches found")
	}

	if len(branchList) == 1 {
		config.ChosenRepo.Branch = branchList[0].Name
		config.ChosenRepo.Sha = branchList[0].Sha
		return config.ChosenRepo, nil
	}

	var prList []apiclient.GitPullRequest
	err = views_util.WithSpinner("Loading", func() error {
		prList, _, err = config.ApiClient.GitProviderAPI.GetRepoPRs(ctx, config.ProviderId, config.NamespaceId, url.QueryEscape(*config.ChosenRepo.Id)).Execute()
		return err
	})

	if err != nil {
		return nil, err
	}

	var branch *apiclient.GitBranch
	if len(prList) == 0 {
		branch = selection.GetBranchFromPrompt(branchList, config.ProjectOrder)
		if branch == nil {
			return nil, common.ErrCtrlCAbort
		}

		config.ChosenRepo.Branch = branch.Name
		config.ChosenRepo.Sha = branch.Sha

		return config.ChosenRepo, nil
	}

	checkoutOptions = append(checkoutOptions, selection.CheckoutDefault)
	checkoutOptions = append(checkoutOptions, selection.CheckoutBranch)
	checkoutOptions = append(checkoutOptions, selection.CheckoutPR)

	chosenCheckoutOption := selection.GetCheckoutOptionFromPrompt(config.ProjectOrder, checkoutOptions)
	if chosenCheckoutOption == selection.CheckoutDefault {
		return config.ChosenRepo, nil
	}

	if chosenCheckoutOption == selection.CheckoutBranch {
		branch = selection.GetBranchFromPrompt(branchList, config.ProjectOrder)
		if branch == nil {
			return nil, common.ErrCtrlCAbort
		}
		config.ChosenRepo.Branch = branch.Name
		config.ChosenRepo.Sha = branch.Sha
	} else if chosenCheckoutOption == selection.CheckoutPR {
		chosenPullRequest := selection.GetPullRequestFromPrompt(prList, config.ProjectOrder)
		if chosenPullRequest == nil {
			return nil, common.ErrCtrlCAbort
		}

		config.ChosenRepo.Branch = chosenPullRequest.Branch
		config.ChosenRepo.Sha = chosenPullRequest.Sha
		config.ChosenRepo.Id = chosenPullRequest.SourceRepoId
		config.ChosenRepo.Name = chosenPullRequest.SourceRepoName
		config.ChosenRepo.Owner = chosenPullRequest.SourceRepoOwner
		config.ChosenRepo.Url = chosenPullRequest.SourceRepoUrl
	}

	return config.ChosenRepo, nil
}
