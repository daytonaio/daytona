// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitproviders

import (
	"fmt"
	"strings"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/server/config"
)

func GetGitProviderForUrl(url string) (gitprovider.GitProvider, error) {
	var gitProvider gitprovider.GitProvider

	c, err := config.GetConfig()
	if err != nil {
		return gitprovider.GitProvider{}, err
	}

	for _, p := range c.GitProviders {
		if strings.Contains(url, fmt.Sprintf("%s.", p.Id)) {
			gitProvider = p
		}

		if p.BaseApiUrl != "" && strings.Contains(url, getHostnameFromUrl(p.BaseApiUrl)) {
			gitProvider = p
		}
	}

	return gitProvider, nil
}

func SetGitProvider(gitProviderData gitprovider.GitProvider) error {
	var providerExists bool

	c, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get config: %s", err.Error())
	}

	for i, provider := range c.GitProviders {
		if provider.Id == gitProviderData.Id {
			c.GitProviders[i].Token = gitProviderData.Token
			c.GitProviders[i].Username = gitProviderData.Username
			c.GitProviders[i].BaseApiUrl = gitProviderData.BaseApiUrl
			providerExists = true
			break
		}
	}

	if !providerExists {
		c.GitProviders = append(c.GitProviders, gitprovider.GitProvider{
			Id:         gitProviderData.Id,
			Token:      gitProviderData.Token,
			Username:   gitProviderData.Username,
			BaseApiUrl: gitProviderData.BaseApiUrl,
		})
	}

	err = config.Save(c)
	if err != nil {
		return fmt.Errorf("failed to save config: %s", err.Error())
	}

	return nil
}

func RemoveGitProvider(gitProviderId string) error {
	c, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get config: %s", err.Error())
	}

	var newProviders []gitprovider.GitProvider
	for _, provider := range c.GitProviders {
		if provider.Id != gitProviderId {
			newProviders = append(newProviders, provider)
		}
	}

	c.GitProviders = newProviders
	err = config.Save(c)
	if err != nil {
		return fmt.Errorf("failed to save config: %s", err.Error())
	}

	return nil
}

func getHostnameFromUrl(url string) string {
	input := url
	input = strings.TrimPrefix(input, "https://")
	input = strings.TrimPrefix(input, "http://")
	input = strings.TrimPrefix(input, "www.")

	// Remove everything after the first '/'
	if slashIndex := strings.Index(input, "/"); slashIndex != -1 {
		input = input[:slashIndex]
	}

	return input
}
