// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitproviders

import (
	"net/url"
	"strconv"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/docker/docker/pkg/stringid"
)

func (s *GitProviderService) GetConfig(id string) (*gitprovider.GitProviderConfig, error) {
	return s.configStore.Find(id)
}

func (s *GitProviderService) ListConfigs() ([]*gitprovider.GitProviderConfig, error) {
	return s.configStore.List()
}

func (s *GitProviderService) ListConfigsForUrl(repoUrl string) ([]*gitprovider.GitProviderConfig, error) {
	var gpcs []*gitprovider.GitProviderConfig

	gitProviders, err := s.configStore.List()
	if err != nil {
		return nil, err
	}

	for _, p := range gitProviders {
		p.Token = url.QueryEscape(p.Token)
		p.Username = url.QueryEscape(p.Username)

		gitProvider, err := s.GetGitProvider(p.Id)
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

func (s *GitProviderService) SetGitProviderConfig(providerConfig *gitprovider.GitProviderConfig) error {
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
		gitProviderConfigs, err := s.ListConfigs()
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

	return s.configStore.Save(providerConfig)
}
