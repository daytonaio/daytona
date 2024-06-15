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
	Id       string  `json:"id"`
	Url      string  `json:"url"`
	Name     string  `json:"name"`
	Branch   *string `json:"branch,omitempty"`
	Sha      *string `json:"sha,omitempty"`
	Owner    string  `json:"owner"`
	PrNumber *uint32 `json:"prNumber,omitempty"`
	Source   string  `json:"source"`
	Path     *string `json:"path,omitempty"`
} // @name StaticGitContext

type GitProvider interface {
	GetNamespaces() ([]*GitNamespace, error)
	GetRepositories(namespace string) ([]*GitRepository, error)
	GetUser() (*GitUser, error)
	GetRepoBranches(repositoryId string, namespaceId string) ([]*GitBranch, error)
	GetRepoPRs(repositoryId string, namespaceId string) ([]*GitPullRequest, error)

	GetRepositoryFromUrl(repositoryUrl string) (*GitRepository, error)
	GetLastCommitSha(staticContext *StaticGitContext) (string, error)
	getPrContext(staticContext *StaticGitContext) (*StaticGitContext, error)
	parseStaticGitContext(repoUrl string) (*StaticGitContext, error)
}

type AbstractGitProvider struct {
	GitProvider
}

func (a *AbstractGitProvider) GetRepositoryFromUrl(repositoryUrl string) (*GitRepository, error) {
	staticContext, err := a.GitProvider.parseStaticGitContext(repositoryUrl)
	if err != nil {
		return nil, fmt.Errorf("Error while parsing git context %s", err.Error())
	}

	if staticContext.PrNumber != nil {
		staticContext, err = a.getPrContext(staticContext)
		if err != nil {
			return nil, fmt.Errorf("Error while fetching PR context %s", err.Error())
		}

	}

	lastCommitSha, err := a.GetLastCommitSha(staticContext)
	if err != nil {
		return nil, fmt.Errorf("Error while fetching last commit SHA %s", err.Error())
	}

	return &GitRepository{
		Id:       staticContext.Id,
		Name:     staticContext.Name,
		Url:      staticContext.Url,
		Branch:   staticContext.Branch,
		Sha:      lastCommitSha,
		Owner:    staticContext.Owner,
		PrNumber: staticContext.PrNumber,
		Source:   staticContext.Source,
		Path:     staticContext.Path,
	}, nil
}

func (a *AbstractGitProvider) parseStaticGitContext(repoUrl string) (*StaticGitContext, error) {
	if strings.HasPrefix(repoUrl, "git@") {
		return a.parseSshGitUrl(repoUrl)
	}

	if !strings.HasPrefix(repoUrl, "http") {
		return nil, errors.New("can not parse git URL: " + repoUrl)
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
	} else {
		defaultBranch := "main"
		repo.Path = &defaultBranch
	}

	repo.Url = getCloneUrl(repo.Source, repo.Owner, repo.Name)
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

	repo.Url = getCloneUrl(repo.Source, repo.Owner, repo.Name)

	return repo, nil
}

func getCloneUrl(source, owner, repo string) string {
	return fmt.Sprintf("https://%s/%s/%s.git", source, owner, repo)
}
