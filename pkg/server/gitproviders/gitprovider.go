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
	"github.com/google/uuid"
)

func (s *GitProviderService) GetGitProviderForUrl(repoUrl string) (gitprovider.GitProvider, string, error) {
	gitProviders, err := s.configStore.List()
	if err != nil {
		return nil, "", err
	}

	for _, p := range gitProviders {
		gitProvider, err := s.GetGitProvider(p.Id)
		if err != nil {
			return nil, "", err
		}

		canHandle, _ := gitProvider.CanHandle(repoUrl)
		if canHandle {
			return gitProvider, p.Id, nil
		}
	}

	u, err := url.Parse(repoUrl)
	if err != nil {
		return nil, "", nil
	}

	hostname := strings.TrimPrefix(u.Hostname(), "www.")
	providerId := strings.Split(hostname, ".")[0]

	gitProvider, err := s.newGitProvider(&gitprovider.GitProviderConfig{
		Id:         providerId,
		Username:   "",
		Token:      "",
		BaseApiUrl: nil,
	})

	return gitProvider, providerId, err
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
		providerConfig.Id = uuid.NewString()
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
