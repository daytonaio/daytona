// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitproviders

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/daytonaio/daytona/pkg/telemetry"
)

type GitProviderServiceConfig struct {
	ConfigStore stores.GitProviderConfigStore

	DetachWorkspaceTemplates func(ctx context.Context, gitProviderConfigId string) error
	TrackTelemetryEvent      func(event telemetry.Event, clientId string) error
}

type GitProviderService struct {
	configStore              stores.GitProviderConfigStore
	detachWorkspaceTemplates func(ctx context.Context, gitProviderConfigId string) error
	trackTelemetryEvent      func(event telemetry.Event, clientId string) error
}

func NewGitProviderService(config GitProviderServiceConfig) services.IGitProviderService {
	return &GitProviderService{
		configStore:              config.ConfigStore,
		detachWorkspaceTemplates: config.DetachWorkspaceTemplates,
		trackTelemetryEvent:      config.TrackTelemetryEvent,
	}
}

var codebergUrl = "https://codeberg.org"

func (s *GitProviderService) GetGitProvider(ctx context.Context, id string) (gitprovider.GitProvider, error) {
	providerConfig, err := s.configStore.Find(ctx, id)
	if err != nil {
		// If config is not defined, use the default (public) client without token
		if stores.IsGitProviderNotFound(err) {
			providerConfig = &models.GitProviderConfig{
				Id:         id,
				ProviderId: id,
				Username:   "",
				Token:      "",
				BaseApiUrl: nil,
			}
		} else {
			return nil, err
		}
	}

	return s.newGitProvider(providerConfig)
}

func (s *GitProviderService) GetLastCommitSha(ctx context.Context, repo *gitprovider.GitRepository) (string, error) {
	var err error
	var provider gitprovider.GitProvider
	providerFound := false

	gitProviders, err := s.configStore.List(ctx)
	if err != nil {
		return "", err
	}

	for _, p := range gitProviders {

		isAwsUrl := strings.Contains(repo.Url, ".amazonaws.com/") || strings.Contains(repo.Url, ".console.aws.amazon.com/")
		if p.ProviderId == "aws-codecommit" && isAwsUrl {
			provider, err = s.GetGitProvider(ctx, p.ProviderId)
			if err == nil {
				return "", err
			}
			providerFound = true
			break
		}

		if strings.Contains(repo.Url, fmt.Sprintf("%s.", p.ProviderId)) {
			provider, err = s.GetGitProvider(ctx, p.ProviderId)
			if err == nil {
				return "", err
			}
			providerFound = true
			break
		}

		hostname, err := getHostnameFromUrl(*p.BaseApiUrl)
		if err != nil {
			return "", err
		}

		if p.BaseApiUrl != nil && strings.Contains(repo.Url, hostname) {
			provider, err = s.GetGitProvider(ctx, p.ProviderId)
			if err == nil {
				return "", err
			}
			providerFound = true
			break
		}

	}

	if !providerFound {
		hostname := strings.TrimPrefix(repo.Source, "www.")
		providerId := strings.Split(hostname, ".")[0]

		provider, err = s.newGitProvider(&models.GitProviderConfig{
			Id:         "",
			ProviderId: providerId,
			Username:   "",
			Token:      "",
			BaseApiUrl: nil,
		})
		if err != nil {
			return "", err
		}
	}

	return provider.GetLastCommitSha(&gitprovider.StaticGitContext{
		Id:       repo.Id,
		Url:      repo.Url,
		Name:     repo.Name,
		Branch:   &repo.Branch,
		Sha:      &repo.Sha,
		Owner:    repo.Owner,
		PrNumber: repo.PrNumber,
		Source:   repo.Source,
		Path:     repo.Path,
	})
}

func (s *GitProviderService) newGitProvider(config *models.GitProviderConfig) (gitprovider.GitProvider, error) {
	baseApiUrl := ""
	if config.BaseApiUrl != nil {
		baseApiUrl = *config.BaseApiUrl
	}

	switch config.ProviderId {
	case "github":
		return gitprovider.NewGitHubGitProvider(config.Token, nil), nil
	case "github-enterprise-server":
		return gitprovider.NewGitHubGitProvider(config.Token, config.BaseApiUrl), nil
	case "gitlab":
		return gitprovider.NewGitLabGitProvider(config.Token, nil), nil
	case "bitbucket":
		return gitprovider.NewBitbucketGitProvider(config.Username, config.Token), nil
	case "bitbucket-server":
		return gitprovider.NewBitbucketServerGitProvider(config.Username, config.Token, baseApiUrl), nil
	case "gitlab-self-managed":
		return gitprovider.NewGitLabGitProvider(config.Token, config.BaseApiUrl), nil
	case "codeberg":
		return gitprovider.NewGiteaGitProvider(config.Token, codebergUrl), nil
	case "gitea":
		return gitprovider.NewGiteaGitProvider(config.Token, baseApiUrl), nil
	case "gitness":
		return gitprovider.NewGitnessGitProvider(config.Token, baseApiUrl), nil
	case "azure-devops":
		return gitprovider.NewAzureDevOpsGitProvider(config.Token, baseApiUrl), nil
	case "aws-codecommit":
		return gitprovider.NewAwsCodeCommitGitProvider(baseApiUrl), nil
	case "gogs":
		return gitprovider.NewGogsGitProvider(config.Token, baseApiUrl), nil
	case "gitee":
		return gitprovider.NewGiteeGitProvider(config.Token), nil
	default:
		return nil, errors.New("git provider not found")
	}
}
