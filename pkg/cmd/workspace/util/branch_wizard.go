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

func SetBranchFromWizard(config BranchWizardConfig) (*apiclient.GitRepository, error) {
	var branchList []apiclient.GitBranch
	var checkoutOptions []selection.CheckoutOption
	page := int32(1)
	perPage := int32(100)
	var err error
	ctx := context.Background()

	for {
		var branches []apiclient.GitBranch
		err = views_util.WithSpinner("Loading", func() error {
			branchList, _, err = config.ApiClient.GitProviderAPI.GetRepoBranches(ctx, config.ProviderId, config.NamespaceId, url.QueryEscape(config.ChosenRepo.Id)).Page(page).PerPage(perPage).Execute()
			return err
		})

		if err != nil {
			return nil, err
		}

		branchList = append(branchList, branches...)

		// Break for git providers with unsupported pagination OR on reaching exhausted items.
		if isGitProviderWithUnsupportedPagination(config.ProviderId) || int32(len(branches)) < perPage {
			break
		}

		page++
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
	perPage = 100

	for {
		var pullRequests []apiclient.GitPullRequest
		err = views_util.WithSpinner("Loading", func() error {
			prList, _, err = config.ApiClient.GitProviderAPI.GetRepoPRs(ctx, config.ProviderId, config.NamespaceId, url.QueryEscape(config.ChosenRepo.Id)).Page(page).PerPage(perPage).Execute()
			return err
		})

		if err != nil {
			return nil, err
		}

		prList = append(prList, pullRequests...)

		// Break for git providers with unsupported pagination OR on reaching exhausted items.
		if isGitProviderWithUnsupportedPagination(config.ProviderId) || int32(len(pullRequests)) < perPage {
			break
		}

		page++
	}

	var branch *apiclient.GitBranch
	namespace := config.Namespace
	if namespace == "" {
		namespace = config.ChosenRepo.Owner
	}

	parentIdentifier := fmt.Sprintf("%s/%s/%s", config.ProviderId, namespace, config.ChosenRepo.Name)
	if len(prList) == 0 {
		branch = selection.GetBranchFromPrompt(branchList, config.ProjectOrder, parentIdentifier)
		if branch == nil {
			return nil, errors.New("must select a branch")
		}

		config.ChosenRepo.Branch = branch.Name
		config.ChosenRepo.Sha = branch.Sha

		return config.ChosenRepo, nil
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
		branch = selection.GetBranchFromPrompt(branchList, config.ProjectOrder, parentIdentifier)
		if branch == nil {
			return nil, errors.New("must select a branch")
		}
		config.ChosenRepo.Branch = branch.Name
		config.ChosenRepo.Sha = branch.Sha
	} else if chosenCheckoutOption == selection.CheckoutPR {
		chosenPullRequest := selection.GetPullRequestFromPrompt(prList, config.ProjectOrder, parentIdentifier)
		if chosenPullRequest == nil {
			return nil, errors.New("must select a pull request")
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
