// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"errors"
	"fmt"
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

type GitProvider interface {
	GetNamespaces() ([]*GitNamespace, error)
	GetRepositories(namespace string) ([]*GitRepository, error)
	GetUser() (*GitUser, error)
	GetRepoBranches(repositoryId string, namespaceId string) ([]*GitBranch, error)
	GetRepoPRs(repositoryId string, namespaceId string) ([]*GitPullRequest, error)

	GetRepositoryFromUrl(repositoryUrl string) (*GitRepository, error)
	GetUrlFromRepository(repository *GitRepository) string
	GetLastCommitSha(staticContext *StaticGitContext) (string, error)
	getPrContext(staticContext *StaticGitContext) (*StaticGitContext, error)
	parseStaticGitContext(repoUrl string) (*StaticGitContext, error)
	GetBranchByCommit(staticContext *StaticGitContext) (string, error)
}

type AbstractGitProvider struct {
	GitProvider
}

func (a *AbstractGitProvider) GetRepositoryFromUrl(repositoryUrl string) (*GitRepository, error) {
	staticContext, err := a.GitProvider.parseStaticGitContext(repositoryUrl)
	if err != nil {
		return nil, err
	}

	if staticContext.PrNumber != nil {
		staticContext, err = a.getPrContext(staticContext)
		if err != nil {
			return nil, err
		}
	}

	var target CloneTarget = CloneTargetBranch
	if staticContext.Branch == staticContext.Sha {
		target = CloneTargetCommit
		branch, err := a.GetBranchByCommit(staticContext)
		if err != nil {
			return nil, err
		}
		*staticContext.Branch = branch
	} else {
		lastCommitSha, err := a.GetLastCommitSha(staticContext)
		if err != nil {
			return nil, err
		}
		*staticContext.Sha = lastCommitSha
	}

	return &GitRepository{
		Id:       staticContext.Id,
		Name:     staticContext.Name,
		Url:      staticContext.Url,
		Branch:   staticContext.Branch,
		Sha:      *staticContext.Sha,
		Owner:    staticContext.Owner,
		PrNumber: staticContext.PrNumber,
		Source:   staticContext.Source,
		Path:     staticContext.Path,
		Target:   target,
	}, nil
}

func (a *AbstractGitProvider) parseStaticGitContext(repoUrl string) (*StaticGitContext, error) {
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
