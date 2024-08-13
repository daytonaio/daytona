// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"errors"
	"net/url"

	"github.com/daytonaio/daytona/pkg/apiclient"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
)

type BranchWizardConfig struct {
	ApiClient    *apiclient.APIClient
	ProviderId   string
	NamespaceId  string
	ChosenRepo   *apiclient.GitRepository
	ProjectOrder int
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
		branchList, _, err = config.ApiClient.GitProviderAPI.GetRepoBranches(ctx, config.ProviderId, config.NamespaceId, url.QueryEscape(config.ChosenRepo.Id)).Execute()
		return err
	})

	if err != nil {
		return nil, err
	}

	if len(branchList) == 0 {
		return nil, errors.New("no branches found")
	}

	if len(branchList) == 1 {
		config.ChosenRepo.Branch = &branchList[0].Name
		config.ChosenRepo.Sha = branchList[0].Sha
		return config.ChosenRepo, nil
	}

	var prList []apiclient.GitPullRequest
	err = views_util.WithSpinner("Loading", func() error {
		prList, _, err = config.ApiClient.GitProviderAPI.GetRepoPRs(ctx, config.ProviderId, config.NamespaceId, url.QueryEscape(config.ChosenRepo.Id)).Execute()
		return err
	})

	if err != nil {
		return nil, err
	}

	var branch *apiclient.GitBranch
	if len(prList) == 0 {
		branch = selection.GetBranchFromPrompt(branchList, config.ProjectOrder)
		if branch == nil {
			return nil, errors.New("must select a branch")
		}

		config.ChosenRepo.Branch = &branch.Name
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
			return nil, errors.New("must select a branch")
		}
		config.ChosenRepo.Branch = &branch.Name
		config.ChosenRepo.Sha = branch.Sha
	} else if chosenCheckoutOption == selection.CheckoutPR {
		chosenPullRequest := selection.GetPullRequestFromPrompt(prList, config.ProjectOrder)
		if chosenPullRequest == nil {
			return nil, errors.New("must select a pull request")
		}

		config.ChosenRepo.Branch = &chosenPullRequest.Branch
		config.ChosenRepo.Sha = chosenPullRequest.Sha
		config.ChosenRepo.Id = chosenPullRequest.SourceRepoId
		config.ChosenRepo.Name = chosenPullRequest.SourceRepoName
		config.ChosenRepo.Owner = chosenPullRequest.SourceRepoOwner
		config.ChosenRepo.Url = chosenPullRequest.SourceRepoUrl
	}

	return config.ChosenRepo, nil
}
