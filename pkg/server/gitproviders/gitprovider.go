// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitproviders

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/docker/docker/pkg/stringid"
)

func (s *GitProviderService) GetGitProviderForUrl(repoUrl string) (gitprovider.GitProvider, string, error) {
	gitProviders, err := s.configStore.List()
	if err != nil {
		return nil, "", err
	}

	for _, p := range gitProviders {
		gitProvider, err := s.GetGitProvider(p.Id)
		if err != nil {
			continue
		}

		canHandle, _ := gitProvider.CanHandle(repoUrl)
		if canHandle {
			_, err = gitProvider.GetRepositoryContext(gitprovider.GetRepositoryContext{
				Url: repoUrl,
			})
			if err == nil {
				return gitProvider, p.Id, nil
			}
		}
	}
	supportedGitProviders := []string{
		"github",
		"github-enterprise-server",
		"gitlab",
		"gitlab-self-managed",
		"bitbucket",
		"bitbucket-server",
		"codeberg",
		"gitea",
		"gitness",
		"azure-devops",
		"aws-codecommit",
	}

	for _, p := range supportedGitProviders {
		gitProvider, err := s.newGitProvider(&gitprovider.GitProviderConfig{
			ProviderId: p,
			Id:         p,
			Username:   "",
			Token:      "",
			BaseApiUrl: nil,
		})
		if err != nil {
			continue
		}
		canHandle, _ := gitProvider.CanHandle(repoUrl)
		if canHandle {
			return gitProvider, p, nil
		}
	}

	return nil, "", errors.New("can not get public client for the URL " + repoUrl)
}

func (s *GitProviderService) GetConfigForUrl(repoUrl string) (*gitprovider.GitProviderConfig, error) {
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
			return p, nil
		}

	}

	supportedGitProviders := config.GetSupportedGitProviders()
	for _, provider := range supportedGitProviders {
		if strings.Contains(repoUrl, provider.Id) {
			return &gitprovider.GitProviderConfig{
				Id: provider.Id,
			}, nil
		}
	}

	return nil, errors.New("git provider not found")
}

func (s *GitProviderService) GetGitProviderForHttpRequest(req *http.Request) (gitprovider.GitProvider, error) {
	var provider *gitprovider.GitProviderConfig

	gitProviders, err := s.configStore.List()
	if err != nil {
		return nil, err
	}

	for _, p := range gitProviders {
		header := req.Header.Get(config.GetWebhookEventHeaderKeyFromGitProvider(p.Id))
		if header == "" {
			continue
		} else {
			provider = p
			break
		}
	}

	if provider == nil {
		return nil, errors.New("git provider for HTTP request not found")
	}

	return s.newGitProvider(provider)
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
		providerConfig.Alias = userData.Username
	}

	return s.configStore.Save(providerConfig)
}

func getHostnameFromUrl(urlToParse string) (string, error) {
	parsed, err := url.Parse(urlToParse)
	if err != nil {
		return "", err
	}

	return strings.TrimPrefix(parsed.Hostname(), "www."), nil
}
