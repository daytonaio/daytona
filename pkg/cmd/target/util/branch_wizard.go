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
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/target/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
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
	disablePagination := false
	curPageItemsNum := 0
	selectionListCursorIdx := 0
	var selectionListOptions views.SelectionListOptions
	var err error

	for {
		err = views_util.WithSpinner("Loading Branches", func() error {
			branches, _, err := config.ApiClient.GitProviderAPI.GetRepoBranches(ctx, config.GitProviderConfigId, url.QueryEscape(config.NamespaceId), url.QueryEscape(config.ChosenRepo.Id)).Page(page).PerPage(perPage).Execute()
			if err != nil {
				return err
			}

			curPageItemsNum = len(branches)
			branchList = append(branchList, branches...)
			return nil
		})

		if err != nil {
			return nil, err
		}

		// Check first if the git provider supports pagination
		if isGitProviderWithUnsupportedPagination(config.GitProviderConfigId) {
			disablePagination = true
		} else {
			// Check if we have reached the end of the list
			disablePagination = int32(curPageItemsNum) < perPage
		}

		selectionListCursorIdx = (int)(page-1) * int(perPage)
		selectionListOptions = views.SelectionListOptions{
			ParentIdentifier:     parentIdentifier,
			IsPaginationDisabled: disablePagination,
			CursorIndex:          selectionListCursorIdx,
		}

		// User will either choose a branch or navigate the pages
		branch, navigate := selection.GetBranchFromPrompt(branchList, config.ProjectOrder, selectionListOptions)
		if !disablePagination && navigate != "" {
			if navigate == views.ListNavigationText {
				page++
				continue // Fetch the next page of branches
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

	err = views_util.WithSpinner("Loading", func() error {
		branches, _, err := config.ApiClient.GitProviderAPI.GetRepoBranches(ctx, config.GitProviderConfigId, url.QueryEscape(config.NamespaceId), url.QueryEscape(config.ChosenRepo.Id)).Page(page).PerPage(perPage).Execute()
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
		prs, _, err := config.ApiClient.GitProviderAPI.GetRepoPRs(ctx, config.GitProviderConfigId, url.QueryEscape(config.NamespaceId), url.QueryEscape(config.ChosenRepo.Id)).Page(page).PerPage(perPage).Execute()
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
		disablePagination := false
		curPageItemsNum := 0
		selectionListCursorIdx := 0
		var selectionListOptions views.SelectionListOptions

		for {
			err = views_util.WithSpinner("Loading Pull Requests", func() error {
				branches, _, err := config.ApiClient.GitProviderAPI.GetRepoBranches(ctx, config.GitProviderConfigId, url.QueryEscape(config.NamespaceId), url.QueryEscape(config.ChosenRepo.Id)).Page(page).PerPage(perPage).Execute()
				if err != nil {
					return err
				}

				curPageItemsNum = len(branches)
				branchList = append(branchList, branches...)
				return nil
			})

			if err != nil {
				return nil, err
			}

			// Check first if the git provider supports pagination
			if isGitProviderWithUnsupportedPagination(config.ProviderId) {
				disablePagination = true
			} else {
				// Check if we have reached the end of the list
				disablePagination = int32(curPageItemsNum) < perPage
			}

			selectionListCursorIdx = (int)(page-1) * int(perPage)
			selectionListOptions = views.SelectionListOptions{
				ParentIdentifier:     parentIdentifier,
				IsPaginationDisabled: disablePagination,
				CursorIndex:          selectionListCursorIdx,
			}

			// User will either choose a PR or navigate the pages
			chosenPullRequest, navigate := selection.GetPullRequestFromPrompt(prList, config.ProjectOrder, selectionListOptions)
			if !disablePagination && navigate != "" {
				if navigate == views.ListNavigationText {
					page++
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
