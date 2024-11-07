// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
)

type BranchWizardParams struct {
	ApiClient           *apiclient.APIClient
	GitProviderConfigId string
	NamespaceId         string
	Namespace           string
	ChosenRepo          *apiclient.GitRepository
	WorkspaceOrder      int
	ProviderId          string
}

func SetBranchFromWizard(params BranchWizardParams) (*apiclient.GitRepository, error) {
	var branchList []apiclient.GitBranch
	var checkoutOptions []selection.CheckoutOption
	page := int32(1)
	// Verify first if num of existing branches is greater than equal to 1
	perPage := int32(2)
	var err error
	ctx := context.Background()

	err = views_util.WithSpinner("Loading", func() error {
		branches, _, err := params.ApiClient.GitProviderAPI.GetRepoBranches(ctx, params.GitProviderConfigId, url.QueryEscape(params.NamespaceId), url.QueryEscape(params.ChosenRepo.Id)).Page(page).PerPage(perPage).Execute()
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
		params.ChosenRepo.Branch = branchList[0].Name
		params.ChosenRepo.Sha = branchList[0].Sha
		return params.ChosenRepo, nil
	}

	var prList []apiclient.GitPullRequest
	page = 1
	// Verify first if there are any existing PRs
	perPage = 1

	err = views_util.WithSpinner("Loading", func() error {
		prs, _, err := params.ApiClient.GitProviderAPI.GetRepoPRs(ctx, params.GitProviderConfigId, url.QueryEscape(params.NamespaceId), url.QueryEscape(params.ChosenRepo.Id)).Page(page).PerPage(perPage).Execute()
		if err != nil {
			return err
		}

		prList = append(prList, prs...)
		return nil
	})

	if err != nil {
		return nil, err
	}

	namespace := params.Namespace
	if namespace == "" {
		namespace = params.ChosenRepo.Owner
	}
	parentIdentifier := fmt.Sprintf("%s/%s/%s", params.ProviderId, namespace, params.ChosenRepo.Name)
	if len(prList) == 0 {
		return runGetBranchFromPromptWithPagination(ctx, params, parentIdentifier, 1, 100)
	}

	checkoutOptions = append(checkoutOptions, selection.CheckoutDefault)
	checkoutOptions = append(checkoutOptions, selection.CheckoutBranch)
	checkoutOptions = append(checkoutOptions, selection.CheckoutPR)

	chosenCheckoutOption := selection.GetCheckoutOptionFromPrompt(params.WorkspaceOrder, checkoutOptions, parentIdentifier)

	if chosenCheckoutOption == (selection.CheckoutOption{}) {
		return nil, common.ErrCtrlCAbort
	}

	if chosenCheckoutOption == selection.CheckoutDefault {
		// Get the default branch from context
		repo, res, err := params.ApiClient.GitProviderAPI.GetGitContext(ctx).Repository(apiclient.GetRepositoryContext{
			Url: params.ChosenRepo.Url,
		}).Execute()
		if err != nil {
			return nil, apiclient_util.HandleErrorResponse(res, err)
		}

		params.ChosenRepo.Branch = repo.Branch

		return params.ChosenRepo, nil
	}

	if chosenCheckoutOption == selection.CheckoutBranch {
		return runGetBranchFromPromptWithPagination(ctx, params, parentIdentifier, 1, 100)
	} else if chosenCheckoutOption == selection.CheckoutPR {
		page = 1
		perPage = 100
		disablePagination := false
		curPageItemsNum := 0
		selectionListCursorIdx := 0
		var selectionListOptions views.SelectionListOptions
		prList = []apiclient.GitPullRequest{}

		for {
			err = views_util.WithSpinner("Loading Pull Requests", func() error {
				prs, _, err := params.ApiClient.GitProviderAPI.GetRepoPRs(ctx, params.GitProviderConfigId, url.QueryEscape(params.NamespaceId), url.QueryEscape(params.ChosenRepo.Id)).Page(page).PerPage(perPage).Execute()
				if err != nil {
					return err
				}

				curPageItemsNum = len(prs)
				prList = append(prList, prs...)
				return nil
			})

			if err != nil {
				return nil, err
			}

			// Check first if the git provider supports pagination
			if isGitProviderWithUnsupportedPagination(params.ProviderId) {
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
			chosenPullRequest, navigate := selection.GetPullRequestFromPrompt(prList, params.WorkspaceOrder, selectionListOptions)
			if !disablePagination && navigate != "" {
				if navigate == views.ListNavigationText {
					page++
					continue
				}
			} else if chosenPullRequest != nil {
				params.ChosenRepo.Branch = chosenPullRequest.Branch
				params.ChosenRepo.Sha = chosenPullRequest.Sha
				params.ChosenRepo.Id = chosenPullRequest.SourceRepoId
				params.ChosenRepo.Name = chosenPullRequest.SourceRepoName
				params.ChosenRepo.Owner = chosenPullRequest.SourceRepoOwner
				params.ChosenRepo.Url = chosenPullRequest.SourceRepoUrl

				return params.ChosenRepo, nil
			} else {
				// If user aborts or there's no selection
				return nil, errors.New("must select a pull request")
			}
		}
	}

	return params.ChosenRepo, nil
}

func runGetBranchFromPromptWithPagination(ctx context.Context, params BranchWizardParams, parentIdentifier string, page, perPage int32) (*apiclient.GitRepository, error) {
	var branchList []apiclient.GitBranch
	disablePagination := false
	curPageItemsNum := 0
	selectionListCursorIdx := 0
	var selectionListOptions views.SelectionListOptions
	var err error

	for {
		err = views_util.WithSpinner("Loading Branches", func() error {
			branches, _, err := params.ApiClient.GitProviderAPI.GetRepoBranches(ctx, params.GitProviderConfigId, url.QueryEscape(params.NamespaceId), url.QueryEscape(params.ChosenRepo.Id)).Page(page).PerPage(perPage).Execute()
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
		if isGitProviderWithUnsupportedPagination(params.GitProviderConfigId) {
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
		branch, navigate := selection.GetBranchFromPrompt(branchList, params.WorkspaceOrder, selectionListOptions)
		if !disablePagination && navigate != "" {
			if navigate == views.ListNavigationText {
				page++
				continue // Fetch the next page of branches
			}
		} else if branch != nil {
			params.ChosenRepo.Branch = branch.Name
			params.ChosenRepo.Sha = branch.Sha

			return params.ChosenRepo, nil
		} else {
			// If user aborts or there's no selection
			return nil, errors.New("must select a branch")
		}
	}
}
