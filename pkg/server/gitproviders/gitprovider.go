// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitproviders

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/models"
)

func (s *GitProviderService) GetGitProviderForUrl(ctx context.Context, repoUrl string) (gitprovider.GitProvider, string, error) {
	gitProviders, err := s.configStore.List(ctx)
	if err != nil {
		return nil, "", err
	}

	var eligibleProvider gitprovider.GitProvider
	var eligibleProviderId string

	for _, p := range gitProviders {
		gitProvider, err := s.GetGitProvider(ctx, p.Id)
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
			eligibleProvider = gitProvider
			eligibleProviderId = p.ProviderId
		}
	}

	if eligibleProvider != nil {
		return eligibleProvider, eligibleProviderId, nil
	}

	for _, p := range config.GetSupportedGitProviders() {
		gitProvider, err := s.newGitProvider(&models.GitProviderConfig{
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

func (s *GitProviderService) GetGitProviderForHttpRequest(ctx context.Context, req *http.Request) (gitprovider.GitProvider, error) {
	var provider *models.GitProviderConfig

	gitProviders, err := s.configStore.List(ctx)
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
