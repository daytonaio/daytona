// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/common"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
)

type BranchWizardConfig struct {
	ApiClient           *apiclient.APIClient
	GitProviderConfigId string
	NamespaceId         string
	Namespace           string
	ChosenRepo          *apiclient.GitRepository
	ProjectOrder        int
	ProviderId          string
}

func runGetBranchFromPromptWithPagination(ctx context.Context, config BranchWizardConfig, parentIdentifier string, page, perPage int32) (*apiclient.GitRepository, error) {
	var branchList []apiclient.GitBranch
	var err error

	for {
		branchList = nil
		err = views_util.WithSpinner("Loading Branches", func() error {
			branches, _, err := config.ApiClient.GitProviderAPI.GetRepoBranches(ctx, config.ProviderId, config.NamespaceId, url.QueryEscape(config.ChosenRepo.Id)).Page(page).PerPage(perPage).Execute()
			if err != nil {
				return err
			}

			branchList = append(branchList, branches...)
			return nil
		})

		if err != nil {
			return nil, err
		}

		// Check if the git provider supports pagination
		isPaginationDisabled := isGitProviderWithUnsupportedPagination(config.ProviderId)

		// User will either choose a branch or navigate the pages
		branch, navigate := selection.GetBranchFromPrompt(branchList, config.ProjectOrder, parentIdentifier, isPaginationDisabled, page, perPage)
		if !isPaginationDisabled && navigate != "" {
			if navigate == "next" {
				page++
				continue // Fetch the next page of branches
			} else if navigate == "prev" && page > 1 {
				page--
				continue // Fetch the previous page of branches
			}
		} else if branch != nil {
			config.ChosenRepo.Branch = branch.Name
			config.ChosenRepo.Sha = branch.Sha

			return config.ChosenRepo, nil
		} else {
			// If user aborts or there's no selection
			return nil, errors.New("must select a branch")
		}
	}
}

func SetBranchFromWizard(config BranchWizardConfig) (*apiclient.GitRepository, error) {
	var branchList []apiclient.GitBranch
	var checkoutOptions []selection.CheckoutOption
	page := int32(1)
	// Verify first if num of existing branches is greater than equal to 1
	perPage := int32(2)
	var err error
	ctx := context.Background()

	err = views_util.WithSpinner("Loading Branches", func() error {
		branches, _, err := config.ApiClient.GitProviderAPI.GetRepoBranches(ctx, config.ProviderId, config.NamespaceId, url.QueryEscape(config.ChosenRepo.Id)).Page(page).PerPage(perPage).Execute()
		if err != nil {
			return err
		}

		branchList = append(branchList, branches...)
		return nil
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
	page = 1
	// Verify first if there are any existing PRs
	perPage = 1

	err = views_util.WithSpinner("Loading", func() error {
		prs, _, err := config.ApiClient.GitProviderAPI.GetRepoPRs(ctx, config.ProviderId, config.NamespaceId, url.QueryEscape(config.ChosenRepo.Id)).Page(page).PerPage(perPage).Execute()
		if err != nil {
			return err
		}

		prList = append(prList, prs...)
		return nil
	})

	if err != nil {
		return nil, err
	}

	namespace := config.Namespace
	if namespace == "" {
		namespace = config.ChosenRepo.Owner
	}
	parentIdentifier := fmt.Sprintf("%s/%s/%s", config.ProviderId, namespace, config.ChosenRepo.Name)
	if len(prList) == 0 {
		return runGetBranchFromPromptWithPagination(ctx, config, parentIdentifier, 1, 100)
	}

	checkoutOptions = append(checkoutOptions, selection.CheckoutDefault)
	checkoutOptions = append(checkoutOptions, selection.CheckoutBranch)
	checkoutOptions = append(checkoutOptions, selection.CheckoutPR)

	chosenCheckoutOption := selection.GetCheckoutOptionFromPrompt(config.ProjectOrder, checkoutOptions, parentIdentifier)

	if chosenCheckoutOption == (selection.CheckoutOption{}) {
		return nil, common.ErrCtrlCAbort
	}

	if chosenCheckoutOption == selection.CheckoutDefault {
		// Get the default branch from context
		repo, res, err := config.ApiClient.GitProviderAPI.GetGitContext(ctx).Repository(apiclient.GetRepositoryContext{
			Url: config.ChosenRepo.Url,
		}).Execute()
		if err != nil {
			return nil, apiclient_util.HandleErrorResponse(res, err)
		}

		config.ChosenRepo.Branch = repo.Branch

		return config.ChosenRepo, nil
	}

	if chosenCheckoutOption == selection.CheckoutBranch {
		return runGetBranchFromPromptWithPagination(ctx, config, parentIdentifier, 1, 100)
	} else if chosenCheckoutOption == selection.CheckoutPR {
		page = 1
		perPage = 100
		for {
			prList = nil
			err = views_util.WithSpinner("Loading Pull Requests", func() error {
				branches, _, err := config.ApiClient.GitProviderAPI.GetRepoBranches(ctx, config.ProviderId, config.NamespaceId, url.QueryEscape(config.ChosenRepo.Id)).Page(page).PerPage(perPage).Execute()
				if err != nil {
					return err
				}

				branchList = append(branchList, branches...)
				return nil
			})

			if err != nil {
				return nil, err
			}

			// Check if the git provider supports pagination
			isPaginationDisabled := isGitProviderWithUnsupportedPagination(config.ProviderId)

			// User will either choose a PR or navigate the pages
			chosenPullRequest, navigate := selection.GetPullRequestFromPrompt(prList, config.ProjectOrder, parentIdentifier, isPaginationDisabled, page, perPage)
			if !isPaginationDisabled && navigate != "" {
				if navigate == "next" {
					page++
					continue
				} else if navigate == "prev" && page > 1 {
					page--
					continue
				}
			} else if chosenPullRequest != nil {
				config.ChosenRepo.Branch = chosenPullRequest.Branch
				config.ChosenRepo.Sha = chosenPullRequest.Sha
				config.ChosenRepo.Id = chosenPullRequest.SourceRepoId
				config.ChosenRepo.Name = chosenPullRequest.SourceRepoName
				config.ChosenRepo.Owner = chosenPullRequest.SourceRepoOwner
				config.ChosenRepo.Url = chosenPullRequest.SourceRepoUrl

				return config.ChosenRepo, nil
			} else {
				// If user aborts or there's no selection
				return nil, errors.New("must select a pull request")
			}
		}
	}

	return config.ChosenRepo, nil
}
