// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"errors"
	"log"
	"net/url"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	gitprovider_view "github.com/daytonaio/daytona/pkg/views/gitprovider"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
)

func isGitProviderWithUnsupportedPagination(providerId string) bool {
	switch providerId {
	case "azure-devops", "bitbucket", "gitness":
		return true
	default:
		return false
	}
}

func getRepositoryFromWizard(userGitProviders []apiclient.GitProvider, additionalProjectOrder int, selectedRepos map[string]int) (*apiclient.GitRepository, error) {
	var providerId string
	var namespaceId string
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
	providerId = selection.GetProviderIdFromPrompt(gitProviderViewList, additionalProjectOrder)
	if providerId == "" {
		return nil, errors.New("must select a provider")
	}

	if providerId == selection.CustomRepoIdentifier {
		return nil, nil
	}

	ctx := context.Background()

	apiClient, err := apiclient_util.GetApiClient(nil)
	if err != nil {
		log.Fatal(err)
	}

	var namespaceList []apiclient.GitNamespace
	page := int32(1)
	perPage := int32(100)

	for {
		var namespaces []apiclient.GitNamespace
		err = views_util.WithSpinner("Loading", func() error {
			namespaces, _, err = apiClient.GitProviderAPI.GetNamespaces(ctx, providerId).Page(page).PerPage(perPage).Execute()
			return err
		})
		if err != nil {
			return nil, err
		}

		namespaceList = append(namespaceList, namespaces...)

		// Break for git providers with unsupported pagination OR on reaching exhausted items.
		if isGitProviderWithUnsupportedPagination(providerId) || int32(len(namespaces)) < perPage {
			break
		}

		page++
	}

	if len(namespaceList) == 1 {
		namespaceId = *namespaceList[0].Id
	} else {
		namespaceId = selection.GetNamespaceIdFromPrompt(namespaceList, additionalProjectOrder)
		if namespaceId == "" {
			return nil, errors.New("namespace not found")
		}
	}

	var providerRepos []apiclient.GitRepository
	page = 1
	perPage = 100

	for {
		var repos []apiclient.GitRepository
		err = views_util.WithSpinner("Loading", func() error {
			repos, _, err = apiClient.GitProviderAPI.GetRepositories(ctx, providerId, namespaceId).Page(page).PerPage(perPage).Execute()
			return err
		})
		if err != nil {
			return nil, err
		}

		providerRepos = append(providerRepos, repos...)

		// For bitbucket, pagination is only supported for fetch repos api
		if providerId == "bitbucket" {
			page++
			continue
		}

		// Break for git providers with unsupported pagination OR on reaching exhausted items.
		if isGitProviderWithUnsupportedPagination(providerId) || int32(len(repos)) < perPage {
			break
		}

		page++
	}

	chosenRepo := selection.GetRepositoryFromPrompt(providerRepos, additionalProjectOrder, selectedRepos)
	if chosenRepo == nil {
		return nil, errors.New("must select a repository")
	}

	var branchList []apiclient.GitBranch
	page = 1
	perPage = 100

	for {
		var branches []apiclient.GitBranch
		err = views_util.WithSpinner("Loading", func() error {
			branches, _, err = apiClient.GitProviderAPI.GetRepoBranches(ctx, providerId, namespaceId, url.QueryEscape(*chosenRepo.Id)).Page(page).PerPage(perPage).Execute()
			return err
		})

		if err != nil {
			return nil, err
		}

		branchList = append(branchList, branches...)

		// Break for git providers with unsupported pagination OR on reaching exhausted items.
		if isGitProviderWithUnsupportedPagination(providerId) || int32(len(branches)) < perPage {
			break
		}

		page++
	}

	if len(branchList) == 0 {
		return nil, errors.New("no branches found")
	}

	if len(branchList) == 1 {
		chosenRepo.Branch = branchList[0].Name
		chosenRepo.Sha = branchList[0].Sha
		return chosenRepo, nil
	}

	var prList []apiclient.GitPullRequest
	page = 1
	perPage = 100

	for {
		var pullRequests []apiclient.GitPullRequest
		err = views_util.WithSpinner("Loading", func() error {
			pullRequests, _, err = apiClient.GitProviderAPI.GetRepoPRs(ctx, providerId, namespaceId, url.QueryEscape(*chosenRepo.Id)).Page(page).PerPage(perPage).Execute()
			return err
		})

		if err != nil {
			return nil, err
		}

		prList = append(prList, pullRequests...)

		// Break for git providers with unsupported pagination OR on reaching exhausted items.
		if isGitProviderWithUnsupportedPagination(providerId) || int32(len(pullRequests)) < perPage {
			break
		}

		page++
	}

	var branch *apiclient.GitBranch
	if len(prList) == 0 {
		branch = selection.GetBranchFromPrompt(branchList, additionalProjectOrder)
		if branch == nil {
			return nil, errors.New("must select a branch")
		}

		chosenRepo.Branch = branch.Name
		chosenRepo.Sha = branch.Sha

		return chosenRepo, nil
	}

	checkoutOptions = append(checkoutOptions, selection.CheckoutDefault)
	checkoutOptions = append(checkoutOptions, selection.CheckoutBranch)
	checkoutOptions = append(checkoutOptions, selection.CheckoutPR)

	chosenCheckoutOption := selection.GetCheckoutOptionFromPrompt(additionalProjectOrder, checkoutOptions)
	if chosenCheckoutOption == selection.CheckoutDefault {
		return chosenRepo, nil
	}

	if chosenCheckoutOption == selection.CheckoutBranch {
		branch = selection.GetBranchFromPrompt(branchList, additionalProjectOrder)
		if branch == nil {
			return nil, errors.New("must select a branch")
		}
		chosenRepo.Branch = branch.Name
		chosenRepo.Sha = branch.Sha
	} else if chosenCheckoutOption == selection.CheckoutPR {
		chosenPullRequest := selection.GetPullRequestFromPrompt(prList, additionalProjectOrder)
		if chosenPullRequest == nil {
			return nil, errors.New("must select a pull request")
		}

		chosenRepo.Branch = chosenPullRequest.Branch
		chosenRepo.Sha = chosenPullRequest.Sha
		chosenRepo.Id = chosenPullRequest.SourceRepoId
		chosenRepo.Name = chosenPullRequest.SourceRepoName
		chosenRepo.Owner = chosenPullRequest.SourceRepoOwner
		chosenRepo.Url = chosenPullRequest.SourceRepoUrl
	}

	return chosenRepo, nil
}
