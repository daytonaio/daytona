// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitproviders

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/gitprovider"
)

func (s *GitProviderService) GetGitProviderForUrl(repoUrl string) (gitprovider.GitProvider, error) {
	gitProviders, err := s.configStore.List()
	if err != nil {
		return nil, err
	}

	for _, p := range gitProviders {
		if strings.Contains(repoUrl, fmt.Sprintf("%s.", p.Id)) {
			return s.GetGitProvider(p.Id)
		}

		if p.BaseApiUrl == nil || *p.BaseApiUrl == "" {
			continue
		}

		hostname, err := getHostnameFromUrl(*p.BaseApiUrl)
		if err != nil {
			return nil, err
		}

		if p.BaseApiUrl != nil && strings.Contains(repoUrl, hostname) {
			return s.GetGitProvider(p.Id)
		}
	}

	u, err := url.Parse(repoUrl)
	if err != nil {
		return nil, err
	}

	hostname := strings.TrimPrefix(u.Hostname(), "www.")
	providerId := strings.Split(hostname, ".")[0]

	return s.newGitProvider(&gitprovider.GitProviderConfig{
		Id:         providerId,
		Username:   "",
		Token:      "",
		BaseApiUrl: nil,
	})
}

func (s *GitProviderService) GetGitProviderForHttpRequest(req *http.Request) (gitprovider.GitProvider, error) {
	var providerId string

	gitProviders, err := s.configStore.List()
	if err != nil {
		return nil, err
	}

	for _, p := range gitProviders {
		header := req.Header.Get(config.GetWebhookEventHeaderKeyFromGitProvider(p.Id))
		if header == "" {
			continue
		} else {
			providerId = p.Id
			break
		}
	}

	return s.newGitProvider(&gitprovider.GitProviderConfig{
		Id:         providerId,
		Username:   "",
		Token:      "",
		BaseApiUrl: nil,
	})
}

func (s *GitProviderService) GetConfigForUrl(url string) (*gitprovider.GitProviderConfig, error) {
	gitProviders, err := s.configStore.List()
	if err != nil {
		return nil, err
	}

	for _, p := range gitProviders {
		if strings.Contains(url, fmt.Sprintf("%s.", p.Id)) {
			return p, nil
		}

		if p.BaseApiUrl == nil || *p.BaseApiUrl == "" {
			continue
		}

		hostname, err := getHostnameFromUrl(*p.BaseApiUrl)
		if err != nil {
			return nil, err
		}

		if p.BaseApiUrl != nil && strings.Contains(url, hostname) {
			return p, nil
		}
	}

	return nil, errors.New("git provider not found")
}

func (s *GitProviderService) SetGitProviderConfig(providerConfig *gitprovider.GitProviderConfig) error {
	gitProvider, err := s.newGitProvider(providerConfig)
	if err != nil {
		return err
	}

	if providerConfig.Username == "" {
		userData, err := gitProvider.GetUser()
		if err != nil {
			return err
		}
		providerConfig.Username = userData.Username
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
