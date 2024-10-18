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
)

func (s *GitProviderService) GetGitProviderForUrl(repoUrl string) (gitprovider.GitProvider, string, error) {
	gitProviders, err := s.configStore.List()
	if err != nil {
		return nil, "", err
	}

	var handleableProviders []string

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
			handleableProviders = append(handleableProviders, p.ProviderId)
		}
	}

	if len(handleableProviders) > 0 {
		providerId := handleableProviders[0]
		gitProvider, err := s.newGitProvider(&gitprovider.GitProviderConfig{
			ProviderId: providerId,
			Id:         providerId,
			Username:   "",
			Token:      "",
			BaseApiUrl: nil,
		})
		if err == nil {
			return gitProvider, providerId, nil
		}
	}

	for _, p := range config.GetSupportedGitProviders() {
		gitProvider, err := s.newGitProvider(&gitprovider.GitProviderConfig{
			ProviderId: p.Id,
			Id:         p.Id,
			Username:   "",
			Token:      "",
			BaseApiUrl: nil,
		})
		if err != nil {
			continue
		}
		canHandle, _ := gitProvider.CanHandle(repoUrl)
		if canHandle {
			return gitProvider, p.Id, nil
		}
	}

	return nil, "", errors.New("can not get public client for the URL " + repoUrl)
}

func (s *GitProviderService) GetGitProviderForHttpRequest(req *http.Request) (gitprovider.GitProvider, error) {
	var provider *gitprovider.GitProviderConfig

	gitProviders, err := s.configStore.List()
	if err != nil {
		return nil, err
	}

	for _, p := range gitProviders {
		header := req.Header.Get(config.GetWebhookEventHeaderKeyFromGitProvider(p.ProviderId))
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

func getHostnameFromUrl(urlToParse string) (string, error) {
	parsed, err := url.Parse(urlToParse)
	if err != nil {
		return "", err
	}

	return strings.TrimPrefix(parsed.Hostname(), "www."), nil
}
