// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"errors"
	"log"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	gitprovider_view "github.com/daytonaio/daytona/pkg/views/gitprovider"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
)

func getRepositoryFromWizard(userGitProviders []serverapiclient.GitProvider, secondaryProjectOrder int) (serverapiclient.GitRepository, error) {
	var providerId string
	var namespaceId string
	var branchName string
	var checkoutOptions []selection.CheckoutOption

	supportedProviders := config.GetSupportedGitProviders()
	var gitProviderViewList []gitprovider_view.GitProviderView

	for _, gitProvider := range userGitProviders {
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
	providerId = selection.GetProviderIdFromPrompt(gitProviderViewList, secondaryProjectOrder)
	if providerId == "" {
		return serverapiclient.GitRepository{}, nil
	}

	if providerId == selection.CustomRepoIdentifier {
		return serverapiclient.GitRepository{
			Id: &selection.CustomRepoIdentifier,
		}, nil
	}

	ctx := context.Background()

	apiClient, err := server.GetApiClient(nil)
	if err != nil {
		log.Fatal(err)
	}

	namespaceList, _, err := apiClient.GitProviderAPI.GetNamespaces(ctx, providerId).Execute()
	if err != nil {
		return serverapiclient.GitRepository{}, err
	}

	if len(namespaceList) == 1 {
		namespaceId = *namespaceList[0].Id
	} else {
		namespaceId = selection.GetNamespaceIdFromPrompt(namespaceList, secondaryProjectOrder)
		if namespaceId == "" {
			return serverapiclient.GitRepository{}, errors.New("namespace not found")
		}
	}

	providerRepos, _, err := apiClient.GitProviderAPI.GetRepositories(ctx, providerId, namespaceId).Execute()
	if err != nil {
		return serverapiclient.GitRepository{}, err
	}

	chosenRepo := selection.GetRepositoryFromPrompt(providerRepos, secondaryProjectOrder)
	if chosenRepo == (serverapiclient.GitRepository{}) {
		return serverapiclient.GitRepository{}, nil
	}

	branchList, _, err := apiClient.GitProviderAPI.GetRepoBranches(ctx, providerId, namespaceId, *chosenRepo.Id).Execute()
	if err != nil {
		return serverapiclient.GitRepository{}, err
	}

	if len(branchList) == 0 {
		return serverapiclient.GitRepository{}, errors.New("no branches found")
	}

	if len(branchList) == 1 {
		branchName = *branchList[0].Name
		chosenRepo.Branch = &branchName
		return chosenRepo, nil
	}

	// TODO: Add support for Bitbucket
	if providerId == "bitbucket" {
		return chosenRepo, nil
	}

	prList, _, err := apiClient.GitProviderAPI.GetRepoPRs(ctx, providerId, namespaceId, *chosenRepo.Id).Execute()
	if err != nil {
		return serverapiclient.GitRepository{}, err
	}
	if len(prList) == 0 {
		branchName = selection.GetBranchNameFromPrompt(branchList, secondaryProjectOrder)
		if branchName == "" {
			return serverapiclient.GitRepository{}, nil
		}
		chosenRepo.Branch = &branchName

		return chosenRepo, nil
	}

	checkoutOptions = append(checkoutOptions, selection.CheckoutDefault)
	checkoutOptions = append(checkoutOptions, selection.CheckoutBranch)
	checkoutOptions = append(checkoutOptions, selection.CheckoutPR)

	chosenCheckoutOption := selection.GetCheckoutOptionFromPrompt(secondaryProjectOrder, checkoutOptions)
	if chosenCheckoutOption == selection.CheckoutDefault {
		return chosenRepo, nil
	}

	if chosenCheckoutOption == selection.CheckoutBranch {
		branchName = selection.GetBranchNameFromPrompt(branchList, secondaryProjectOrder)
		if branchName == "" {
			return serverapiclient.GitRepository{}, nil
		}
		chosenRepo.Branch = &branchName
	} else if chosenCheckoutOption == selection.CheckoutPR {
		chosenPullRequest := selection.GetPullRequestFromPrompt(prList, secondaryProjectOrder)
		if *chosenPullRequest.Branch == "" {
			return serverapiclient.GitRepository{}, nil
		}

		chosenRepo.Branch = chosenPullRequest.Branch
	}

	return chosenRepo, nil
}
