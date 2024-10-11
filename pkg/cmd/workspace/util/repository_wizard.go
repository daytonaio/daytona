// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"fmt"

	config_const "github.com/daytonaio/daytona/cmd/daytona/config"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/views"
	gitprovider_view "github.com/daytonaio/daytona/pkg/views/gitprovider"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/create"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"

	log "github.com/sirupsen/logrus"
)

func isGitProviderWithUnsupportedPagination(providerId string) bool {
	switch providerId {
	case "azure-devops", "bitbucket", "gitness", "aws-codecommit":
		return true
	default:
		return false
	}
}

func gitProviderAppendsPersonalNamespace(providerId string) bool {
	switch providerId {
	case "github", "gitlab", "gitea":
		return true
	default:
		return false
	}
}

type RepositoryWizardConfig struct {
	ApiClient           *apiclient.APIClient
	UserGitProviders    []apiclient.GitProvider
	Manual              bool
	MultiProject        bool
	SkipBranchSelection bool
	ProjectOrder        int
	SelectedRepos       map[string]int
}

func getRepositoryFromWizard(config RepositoryWizardConfig) (*apiclient.GitRepository, string, error) {
	var gitProviderConfigId string
	var namespaceId string
	var err error

	ctx := context.Background()

	samples, res, err := config.ApiClient.SampleAPI.ListSamples(ctx).Execute()
	if err != nil {
		log.Debug("Error fetching samples: ", apiclient_util.HandleErrorResponse(res, err))
	}

	if (len(config.UserGitProviders) == 0 && len(samples) == 0) || config.Manual {
		repo, err := create.GetRepositoryFromUrlInput(config.MultiProject, config.ProjectOrder, config.ApiClient, config.SelectedRepos)
		return repo, selection.CustomRepoIdentifier, err
	}

	supportedProviders := config_const.GetSupportedGitProviders()
	var gitProviderViewList []gitprovider_view.GitProviderView

	for _, gitProvider := range config.UserGitProviders {
		for _, supportedProvider := range supportedProviders {
			if gitProvider.ProviderId == supportedProvider.Id {
				gitProviderViewList = append(gitProviderViewList,
					gitprovider_view.GitProviderView{
						Id:         gitProvider.Id,
						ProviderId: gitProvider.ProviderId,
						Name:       supportedProvider.Name,
						Username:   gitProvider.Username,
						Alias:      gitProvider.Alias,
					},
				)
			}
		}
	}

	gitProviderConfigId = selection.GetProviderIdFromPrompt(gitProviderViewList, config.ProjectOrder, len(samples) > 0)
	if gitProviderConfigId == "" {
		return nil, "", common.ErrCtrlCAbort
	}

	if gitProviderConfigId == selection.CustomRepoIdentifier {
		repo, err := create.GetRepositoryFromUrlInput(config.MultiProject, config.ProjectOrder, config.ApiClient, config.SelectedRepos)
		return repo, selection.CustomRepoIdentifier, err
	}

	if gitProviderConfigId == selection.CREATE_FROM_SAMPLE {
		sample := selection.GetSampleFromPrompt(samples)
		if sample == nil {
			return nil, "", common.ErrCtrlCAbort
		}

		repo, res, err := config.ApiClient.GitProviderAPI.GetGitContext(ctx).Repository(apiclient.GetRepositoryContext{
			Url: sample.GitUrl,
		}).Execute()
		if err != nil {
			return nil, "", apiclient_util.HandleErrorResponse(res, err)
		}

		return repo, selection.CREATE_FROM_SAMPLE, nil
	}

	var providerId string
	for _, gp := range gitProviderViewList {
		if gp.Id == gitProviderConfigId {
			providerId = gp.ProviderId
		}
	}

	var navigate string
	page := int32(1)
	perPage := int32(100)
	disablePagination := false
	curPageItemsNum := 0
	selectionListCursorIdx := 0
	var selectionListOptions views.SelectionListOptions

	var namespaceList []apiclient.GitNamespace
	namespace := ""

	for {
		err = views_util.WithSpinner("Loading Namespaces", func() error {
			namespaces, _, err := config.ApiClient.GitProviderAPI.GetNamespaces(ctx, gitProviderConfigId).Page(page).PerPage(perPage).Execute()
			if err != nil {
				return err
			}
			curPageItemsNum = len(namespaces)
			namespaceList = append(namespaceList, namespaces...)
			return nil
		})

		if err != nil {
			return nil, "", err
		}

		if len(namespaceList) == 1 {
			namespaceId = namespaceList[0].Id
			namespace = namespaceList[0].Name
			break
		}

		// Check first if the git provider supports pagination
		if isGitProviderWithUnsupportedPagination(providerId) {
			disablePagination = true
		} else {
			// Check if we have reached the end of the list
			// For few providers, we manually append "personal" namespace info on the first page
			if page == 1 && gitProviderAppendsPersonalNamespace(providerId) {
				disablePagination = int32(curPageItemsNum)-1 < perPage
			} else {
				disablePagination = int32(curPageItemsNum) < perPage
			}
		}

		selectionListCursorIdx = (int)(page-1) * int(perPage)
		selectionListOptions = views.SelectionListOptions{
			ParentIdentifier:     providerId,
			IsPaginationDisabled: disablePagination,
			CursorIndex:          selectionListCursorIdx,
		}

		namespaceId, navigate = selection.GetNamespaceIdFromPrompt(namespaceList, config.ProjectOrder, selectionListOptions)

		if !disablePagination && navigate != "" {
			if navigate == views.ListNavigationText {
				page++
				continue
			}
		} else if namespaceId != "" {
			for _, namespaceItem := range namespaceList {
				if namespaceItem.Id == namespaceId {
					namespace = namespaceItem.Name
				}
			}
			break
		} else {
			// If user aborts or there's no selection
			return nil, "", common.ErrCtrlCAbort
		}
	}

	var providerRepos []apiclient.GitRepository
	var chosenRepo *apiclient.GitRepository
	page = 1
	perPage = 100

	parentIdentifier := fmt.Sprintf("%s/%s", providerId, namespace)
	for {
		// Fetch repos for the current page
		err = views_util.WithSpinner("Loading Repositories", func() error {

			repos, _, err := config.ApiClient.GitProviderAPI.GetRepositories(ctx, gitProviderConfigId, namespaceId).Page(page).PerPage(perPage).Execute()
			if err != nil {
				return err
			}
			curPageItemsNum = len(repos)
			providerRepos = append(providerRepos, repos...)
			return nil
		})

		if err != nil {
			return nil, "", err
		}

		// Check first if the git provider supports pagination
		// For bitbucket, pagination is only supported for GET repos api, Not for its' GET branches/ namespaces/ pull-requests apis.
		if isGitProviderWithUnsupportedPagination(providerId) && providerId != "bitbucket" {
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

		// User will either choose a repo or navigate the pages
		chosenRepo, navigate = selection.GetRepositoryFromPrompt(providerRepos, config.ProjectOrder, config.SelectedRepos, selectionListOptions)
		if !disablePagination && navigate != "" {
			if navigate == views.ListNavigationText {
				page++
				continue // Fetch the next page of repos
			}
		} else if chosenRepo != nil {
			break
		} else {
			// If user aborts or there's no selection
			return nil, "", common.ErrCtrlCAbort
		}
	}

	if config.SkipBranchSelection {
		return chosenRepo, gitProviderConfigId, nil
	}

	repoWithBranch, err := SetBranchFromWizard(BranchWizardConfig{
		ApiClient:           config.ApiClient,
		GitProviderConfigId: gitProviderConfigId,
		NamespaceId:         namespaceId,
		Namespace:           namespace,
		ChosenRepo:          chosenRepo,
		ProjectOrder:        config.ProjectOrder,
		ProviderId:          providerId,
	})

	return repoWithBranch, gitProviderConfigId, err
}
