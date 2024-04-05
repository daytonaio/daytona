// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitproviders

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/daytonaio/daytona/pkg/gitprovider"
)

func (s *GitProviderService) GetGitProviderForUrl(url string) (gitprovider.GitProvider, error) {
	gitProviders, err := s.configStore.List()
	if err != nil {
		return nil, err
	}

	for _, p := range gitProviders {
		if strings.Contains(url, fmt.Sprintf("%s.", p.Id)) {
			return s.GetGitProvider(p.Id)
		}

		hostname, err := getHostnameFromUrl(*p.BaseApiUrl)
		if err != nil {
			return nil, err
		}

		if p.BaseApiUrl != nil && strings.Contains(url, hostname) {
			return s.GetGitProvider(p.Id)
		}
	}

	return nil, errors.New("git provider not found")
}

func (s *GitProviderService) SetGitProviderConfig(providerConfig *gitprovider.GitProviderConfig) error {
	return s.configStore.Save(providerConfig)
}

func getHostnameFromUrl(urlToParse string) (string, error) {
	parsed, err := url.Parse(urlToParse)
	if err != nil {
		return "", err
	}

	return strings.TrimPrefix("www.", parsed.Hostname()), nil
}
