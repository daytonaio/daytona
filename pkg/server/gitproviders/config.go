// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitproviders

import (
	"context"
	"net/url"
	"strconv"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/docker/docker/pkg/stringid"
)

func (s *GitProviderService) GetConfig(ctx context.Context, id string) (*models.GitProviderConfig, error) {
	return s.configStore.Find(ctx, id)
}

func (s *GitProviderService) ListConfigs(ctx context.Context) ([]*models.GitProviderConfig, error) {
	return s.configStore.List(ctx)
}

func (s *GitProviderService) ListConfigsForUrl(ctx context.Context, repoUrl string) ([]*models.GitProviderConfig, error) {
	var gpcs []*models.GitProviderConfig

	gitProviders, err := s.configStore.List(ctx)
	if err != nil {
		return nil, err
	}

	for _, p := range gitProviders {
		p.Token = url.QueryEscape(p.Token)
		p.Username = url.QueryEscape(p.Username)

		gitProvider, err := s.GetGitProvider(ctx, p.Id)
		if err != nil {
			return nil, err
		}

		canHandle, _ := gitProvider.CanHandle(repoUrl)
		if canHandle {
			_, err = gitProvider.GetRepositoryContext(gitprovider.GetRepositoryContext{
				Url: repoUrl,
			})
			if err == nil {
				gpcs = append(gpcs, p)
			}
		}
	}

	return gpcs, nil
}

func (s *GitProviderService) SetGitProviderConfig(ctx context.Context, providerConfig *models.GitProviderConfig) error {
	gitProvider, err := s.newGitProvider(providerConfig)
	if err != nil {
		return err
	}

	userData, err := gitProvider.GetUser()
	if err != nil {
		return err
	}
	providerConfig.Username = userData.Username
	if providerConfig.Id == "" {
		id := stringid.GenerateRandomID()
		id = stringid.TruncateID(id)
		providerConfig.Id = id
	}

	if providerConfig.Alias == "" {
		gitProviderConfigs, err := s.ListConfigs(ctx)
		if err != nil {
			return err
		}

		uniqueAlias := userData.Username
		aliases := make(map[string]bool)

		for _, c := range gitProviderConfigs {
			aliases[c.Alias] = true
		}
		counter := 2

		for aliases[uniqueAlias] {
			uniqueAlias = userData.Username + strconv.Itoa(counter)
			counter++
		}

		providerConfig.Alias = uniqueAlias
	}

	return s.configStore.Save(ctx, providerConfig)
}
