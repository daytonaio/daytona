// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const personalNamespaceId = "<PERSONAL>"

type StaticGitContext struct {
	Id       string  `json:"id" validate:"required"`
	Url      string  `json:"url" validate:"required"`
	Name     string  `json:"name" validate:"required"`
	Branch   *string `json:"branch,omitempty" validate:"optional"`
	Sha      *string `json:"sha,omitempty" validate:"optional"`
	Owner    string  `json:"owner" validate:"required"`
	PrNumber *uint32 `json:"prNumber,omitempty" validate:"optional"`
	Source   string  `json:"source" validate:"required"`
	Path     *string `json:"path,omitempty" validate:"optional"`
} // @name StaticGitContext

type GetRepositoryContext struct {
	Id       *string `json:"id" validate:"optional"`
	Url      string  `json:"url" validate:"required"`
	Name     *string `json:"name" validate:"optional"`
	Branch   *string `json:"branch,omitempty" validate:"optional"`
	Sha      *string `json:"sha" validate:"optional"`
	Owner    *string `json:"owner" validate:"optional"`
	PrNumber *uint32 `json:"prNumber,omitempty" validate:"optional"`
	Source   *string `json:"source" validate:"optional"`
	Path     *string `json:"path,omitempty" validate:"optional"`
} // @name GetRepositoryContext

// ListOptions holds additional parameters for api responses.
type ListOptions struct {
	Page    int `json:"page,omitempty"`
	PerPage int `json:"perPage,omitempty"`
}

type GitProvider interface {
	GetNamespaces(options ListOptions) ([]*GitNamespace, error)
	GetRepositories(namespace string, options ListOptions) ([]*GitRepository, error)
	GetUser() (*GitUser, error)
	GetRepoBranches(repositoryId string, namespaceId string, options ListOptions) ([]*GitBranch, error)
	GetRepoPRs(repositoryId string, namespaceId string, options ListOptions) ([]*GitPullRequest, error)

	CanHandle(repoUrl string) (bool, error)
	GetRepositoryContext(repoContext GetRepositoryContext) (*GitRepository, error)
	GetUrlFromContext(repoContext *GetRepositoryContext) string
	GetLastCommitSha(staticContext *StaticGitContext) (string, error)
	GetBranchByCommit(staticContext *StaticGitContext) (string, error)
	GetPrContext(staticContext *StaticGitContext) (*StaticGitContext, error)
	ParseStaticGitContext(repoUrl string) (*StaticGitContext, error)
	GetDefaultBranch(staticContext *StaticGitContext) (*string, error)

	RegisterPrebuildWebhook(repo *GitRepository, endpointUrl string) (string, error)
	GetPrebuildWebhook(repo *GitRepository, endpointUrl string) (*string, error)
	UnregisterPrebuildWebhook(repo *GitRepository, id string) error
	GetCommitsRange(repo *GitRepository, initialSha string, currentSha string) (int, error)
	ParseEventData(request *http.Request) (*GitEventData, error)
}

type AbstractGitProvider struct {
	GitProvider
}

func (a *AbstractGitProvider) GetRepositoryContext(repoContext GetRepositoryContext) (*GitRepository, error) {
	staticContext, err := a.GitProvider.ParseStaticGitContext(repoContext.Url)
	if err != nil {
		return nil, err
	}

	if repoContext.PrNumber != nil {
		staticContext.PrNumber = repoContext.PrNumber
		staticContext, err = a.GetPrContext(staticContext)
		if err != nil {
			return nil, err
		}
	} else {
		if staticContext.PrNumber != nil {
			staticContext, err = a.GetPrContext(staticContext)
			if err != nil {
				return nil, err
			}
		}
	}

	if repoContext.Branch != nil {
		staticContext.Branch = repoContext.Branch
	}

	var target CloneTarget = CloneTargetBranch
	if staticContext.Sha != nil && staticContext.Branch == staticContext.Sha {
		target = CloneTargetCommit
		branch, err := a.GetBranchByCommit(staticContext)
		if err != nil {
			return nil, err
		}
		staticContext.Branch = &branch
	} else {
		if staticContext.Branch == nil {
			branch, err := a.GitProvider.GetDefaultBranch(staticContext)
			if err != nil {
				return nil, err
			}
			staticContext.Branch = branch
		}
		lastCommitSha, err := a.GetLastCommitSha(staticContext)
		if err != nil {
			return nil, err
		}
		staticContext.Sha = &lastCommitSha
	}

	return &GitRepository{
		Id:       staticContext.Id,
		Name:     staticContext.Name,
		Url:      staticContext.Url,
		Branch:   *staticContext.Branch,
		Sha:      *staticContext.Sha,
		Owner:    staticContext.Owner,
		PrNumber: staticContext.PrNumber,
		Source:   staticContext.Source,
		Path:     staticContext.Path,
		Target:   target,
	}, nil
}

func (a *AbstractGitProvider) ParseStaticGitContext(repoUrl string) (*StaticGitContext, error) {
	isHttps := true
	if strings.HasPrefix(repoUrl, "http://") {
		isHttps = false
	} else if strings.HasPrefix(repoUrl, "git@") {
		return a.parseSshGitUrl(repoUrl)
	} else if !strings.HasPrefix(repoUrl, "http") {
		return nil, errors.New("cannot parse git URL: " + repoUrl)
	}

	repoUrl = strings.TrimSuffix(repoUrl, ".git")

	u, err := url.Parse(repoUrl)
	if err != nil {
		return nil, err
	}

	repo := &StaticGitContext{}

	path := strings.TrimPrefix(u.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		return nil, errors.New("cannot parse git URL: " + repoUrl)
	}

	repo.Source = u.Host
	repo.Owner = parts[0]
	repo.Name = parts[1]
	repo.Id = parts[1]

	branchPath := strings.Join(parts[2:], "/")
	if branchPath != "" {
		repo.Path = &branchPath
	}

	repo.Url = getCloneUrl(repo.Source, repo.Owner, repo.Name, isHttps)

	return repo, nil
}

func (g *AbstractGitProvider) GetPrebuildWebhook(repo *GitRepository, endpointUrl string) (*string, error) {
	return nil, errors.New("prebuilds not yet implemented for this git provider")
}

func (g *AbstractGitProvider) RegisterPrebuildWebhook(repo *GitRepository, endpointUrl string) (string, error) {
	return "", errors.New("prebuilds not yet implemented for this git provider")
}

func (g *AbstractGitProvider) UnregisterPrebuildWebhook(repo *GitRepository, id string) error {
	return errors.New("prebuilds not yet implemented for this git provider")
}

func (g *AbstractGitProvider) GetCommitsRange(repo *GitRepository, initialSha string, currentSha string) (int, error) {
	return 0, errors.New("prebuilds not yet implemented for this git provider")
}

func (g *AbstractGitProvider) ParseEventData(request *http.Request) (*GitEventData, error) {
	return nil, errors.New("prebuilds not yet implemented for this git provider")
}

func (a *AbstractGitProvider) parseSshGitUrl(gitURL string) (*StaticGitContext, error) {
	re := regexp.MustCompile(`git@([\w\.]+):(.+?)/(.+?)(?:\.git)?$`)
	matches := re.FindStringSubmatch(gitURL)
	if len(matches) != 4 {
		return nil, errors.New("cannot parse git URL: " + gitURL)
	}

	repo := &StaticGitContext{}

	repo.Source = matches[1]
	repo.Owner = matches[2]
	repo.Name = matches[3]
	repo.Id = matches[3]

	repo.Url = getCloneUrl(repo.Source, repo.Owner, repo.Name, true)

	return repo, nil
}

func getCloneUrl(source, owner, repo string, isHttps bool) string {
	scheme := "http"
	if isHttps {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s.git", scheme, source, owner, repo)
}
