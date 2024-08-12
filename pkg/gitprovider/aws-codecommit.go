// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/codecommit"
	"github.com/aws/aws-sdk-go-v2/service/codecommit/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go/aws"
	log "github.com/sirupsen/logrus"
)

type AwsCodeCommitGitProvider struct {
	*AbstractGitProvider
	baseApiUrl string
	region     string
}

func NewAwsCodeCommitGitProvider(baseApiUrl string) *AwsCodeCommitGitProvider {
	region := strings.TrimPrefix(strings.Split(baseApiUrl, ".")[0], "https://")
	gitProvider := &AwsCodeCommitGitProvider{
		AbstractGitProvider: &AbstractGitProvider{},
		baseApiUrl:          baseApiUrl,
		region:              region,
	}
	gitProvider.AbstractGitProvider.GitProvider = gitProvider

	return gitProvider
}

func (g *AwsCodeCommitGitProvider) GetNamespaces() ([]*GitNamespace, error) {
	// AWS CodeCommit does not have a project and repository structure similar to other git providers.
	// Therefore, returning repositories as an array of type GitNamespace.
	client, err := g.getApiClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %s", err.Error())
	}
	repositories, err := client.ListRepositories(context.TODO(), &codecommit.ListRepositoriesInput{
		SortBy: types.SortByEnumRepositoryName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repositories: %s", err.Error())
	}
	var namespaces []*GitNamespace
	for _, repository := range repositories.Repositories {
		namespace := &GitNamespace{
			Id:   *repository.RepositoryName,
			Name: *repository.RepositoryName,
		}
		namespaces = append(namespaces, namespace)
	}

	return namespaces, nil
}

func (g *AwsCodeCommitGitProvider) GetUrlFromRepository(repository *GitRepository) string {
	baseURL := ""
	if strings.Contains(repository.Source, "git-codecommit") {
		region := strings.Split(repository.Source, ".")[1]
		baseURL = fmt.Sprintf("https://%s.console.aws.amazon.com/codesuite/codecommit/repositories/%s", region, repository.Name)
	} else {
		baseURL = fmt.Sprintf("https://%s/codesuite/codecommit/repositories/%s", repository.Source, repository.Name)
	}

	if repository.Sha != "" {
		return fmt.Sprintf("%s/commit/%s", baseURL, repository.Sha)
	}

	if repository.PrNumber != nil {
		return fmt.Sprintf("%s/pull-requests/%d", baseURL, *repository.PrNumber)
	}

	if repository.Branch != nil && *repository.Branch != "" {
		branchURL := fmt.Sprintf("%s/browse/refs/heads/%s", baseURL, *repository.Branch)

		if repository.Path != nil && *repository.Path != "" {
			return fmt.Sprintf("%s/--/%s", branchURL, *repository.Path)
		}
		return branchURL
	}

	if repository.Path != nil && *repository.Path != "" {
		return fmt.Sprintf("%s/browse/refs/heads/main/--/%s", baseURL, *repository.Path)
	}

	return fmt.Sprintf("%s/browse", baseURL)

}

func (g *AwsCodeCommitGitProvider) getApiClient() (*codecommit.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS SDK configuration")
	}
	client := codecommit.NewFromConfig(cfg)

	return client, nil
}

func (g *AwsCodeCommitGitProvider) GetRepositories(namespace string) ([]*GitRepository, error) {
	client, err := g.getApiClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %s", err.Error())
	}
	var repos []*GitRepository
	data, err := client.GetRepository(context.TODO(), &codecommit.GetRepositoryInput{
		RepositoryName: &namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository: %s with error: %v", namespace, err.Error())
	}
	repository := &GitRepository{
		Id:     *data.RepositoryMetadata.RepositoryName,
		Name:   *data.RepositoryMetadata.RepositoryName,
		Url:    getCodeCommitCloneUrl(g.region, *data.RepositoryMetadata.RepositoryName),
		Branch: data.RepositoryMetadata.DefaultBranch,
		Owner:  *data.RepositoryMetadata.AccountId,
	}
	modifiedURLstring := strings.Replace(*data.RepositoryMetadata.CloneUrlHttp, "git-codecommit.", "", 1)
	modifiedurl, err := url.Parse(modifiedURLstring)
	if err != nil {
		log.Warningf("failed to extract source of repository: %s", *data.RepositoryMetadata.RepositoryName)
	} else {
		repository.Source = modifiedurl.Host
	}
	repos = append(repos, repository)

	return repos, nil
}

func (g *AwsCodeCommitGitProvider) GetRepoBranches(repositoryId string, namespaceId string) ([]*GitBranch, error) {
	client, err := g.getApiClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %s", err.Error())
	}
	branchesoutput, err := client.ListBranches(context.TODO(), &codecommit.ListBranchesInput{
		RepositoryName: &repositoryId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch branches for repository: %s", repositoryId)
	}
	var gitbranches []*GitBranch
	for _, branch := range branchesoutput.Branches {
		var gitbranch *GitBranch

		br, err := client.GetBranch(context.TODO(), &codecommit.GetBranchInput{
			BranchName:     aws.String(branch),
			RepositoryName: aws.String(repositoryId),
		})
		if err != nil {
			gitbranch = &GitBranch{
				Name: branch,
			}
		} else {
			gitbranch = &GitBranch{
				Name: *br.Branch.BranchName,
				Sha:  *br.Branch.CommitId,
			}
		}
		gitbranches = append(gitbranches, gitbranch)
	}
	return gitbranches, nil
}

func (g *AwsCodeCommitGitProvider) GetUser() (*GitUser, error) {
	// AWS CodeCommit does not provide an API to get user details directly.
	// Therefore, we are using the IAM service API to get user details.
	// No extra configuration is needed for the IAM service API.
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS SDK configuration")
	}
	iamclient := iam.NewFromConfig(cfg)
	user, err := iamclient.GetUser(context.TODO(), &iam.GetUserInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %s", err.Error())
	}

	// IAM service does not provide the email in the GetUser API response.
	// Therefore, the email field is set to an empty string.
	// TODO: Update the email field if IAM service starts providing the email in the GetUser API.
	return &GitUser{
		Id:       *user.User.UserId,
		Username: *user.User.UserName,
		Name:     *user.User.Arn,
		Email:    "",
	}, nil

}

func (g *AwsCodeCommitGitProvider) GetRepoPRs(repositoryId string, namespaceId string) ([]*GitPullRequest, error) {
	client, err := g.getApiClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %s", err.Error())
	}
	pullrequests, err := client.ListPullRequests(context.TODO(), &codecommit.ListPullRequestsInput{
		RepositoryName:    &repositoryId,
		PullRequestStatus: types.PullRequestStatusEnumOpen,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get pull requests: %s", err.Error())
	}
	var pullRequests []*GitPullRequest
	for _, pullrequestid := range pullrequests.PullRequestIds {
		pr, err := client.GetPullRequest(context.TODO(), &codecommit.GetPullRequestInput{
			PullRequestId: &pullrequestid,
		})
		if err != nil {
			log.Warnf("failed to get pull request: %s error: %f", pullrequestid, err)
		}
		prbranch := strings.TrimPrefix(*pr.PullRequest.PullRequestTargets[0].SourceReference, "refs/heads/")
		pullRequest := &GitPullRequest{
			Name:            *pr.PullRequest.Title,
			Branch:          *pr.PullRequest.PullRequestTargets[0].SourceReference,
			Sha:             *pr.PullRequest.PullRequestTargets[0].SourceReference,
			SourceRepoId:    prbranch,
			SourceRepoName:  prbranch,
			SourceRepoUrl:   getCodeCommitCloneUrl(g.region, repositoryId),
			SourceRepoOwner: strings.Split(*pr.PullRequest.AuthorArn, ":")[5],
		}
		pullRequests = append(pullRequests, pullRequest)
	}
	return pullRequests, nil
}

func (g *AwsCodeCommitGitProvider) GetLastCommitSha(staticContext *StaticGitContext) (string, error) {
	client, err := g.getApiClient()
	if err != nil {
		return "", fmt.Errorf("failed to get client: %s", err.Error())
	}
	branch, err := client.GetBranch(context.TODO(), &codecommit.GetBranchInput{
		RepositoryName: &staticContext.Name,
		BranchName:     staticContext.Branch,
	})
	if err != nil {
		return "", fmt.Errorf("failed to fetch branch details: %w", err)
	}
	return *branch.Branch.CommitId, nil
}

func (g *AwsCodeCommitGitProvider) getPrContext(staticContext *StaticGitContext) (*StaticGitContext, error) {
	if staticContext.PrNumber == nil {
		return staticContext, nil
	}
	client, err := g.getApiClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %s", err.Error())
	}
	prnumber := strconv.FormatUint(uint64(*staticContext.PrNumber), 10)
	pr, err := client.GetPullRequest(context.TODO(), &codecommit.GetPullRequestInput{
		PullRequestId: aws.String(prnumber),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get pull request: %s", err.Error())
	}
	repo := *staticContext
	prbranchname := strings.TrimPrefix(*pr.PullRequest.PullRequestTargets[0].SourceReference, "refs/heads/")
	repo.Branch = &prbranchname
	return &repo, nil
}

func (g *AwsCodeCommitGitProvider) GetBranchByCommit(staticContext *StaticGitContext) (string, error) {
	if staticContext.Sha == nil || *staticContext.Sha == "" {
		return *staticContext.Branch, nil
	}

	client, err := g.getApiClient()
	if err != nil {
		return "", err
	}

	branches, err := client.ListBranches(context.TODO(), &codecommit.ListBranchesInput{
		RepositoryName: aws.String(staticContext.Name),
	})
	if err != nil {
		return "", fmt.Errorf("failed to list branches, %v", err)
	}

	var branchName string
	for _, branch := range branches.Branches {
		branchInfo, err := client.GetBranch(context.TODO(), &codecommit.GetBranchInput{
			RepositoryName: aws.String(staticContext.Name),
			BranchName:     aws.String(branch),
		})
		if err != nil {
			log.Infof("failed to get branch info for %s, %v", branch, err)
			continue
		}

		if branchInfo.Branch.CommitId != nil && *branchInfo.Branch.CommitId == *staticContext.Sha {
			branchName = branch
			break
		}
	}

	if branchName == "" {
		return "", fmt.Errorf("branch not found for SHA: %s", *staticContext.Sha)
	}

	return branchName, nil
}

func (g *AwsCodeCommitGitProvider) parseStaticGitContext(repoUrl string) (*StaticGitContext, error) {
	url, err := url.Parse(repoUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %s with error: %v", repoUrl, err.Error())
	}
	branch := "main"
	if strings.Contains(repoUrl, "git-codecommit") {
		reponame := strings.TrimPrefix(url.Path, "/v1/repos/")
		branchpath := fmt.Sprintf("refs/heads/%s/", branch)
		return &StaticGitContext{
			Id:       reponame,
			Name:     reponame,
			Owner:    reponame,
			Source:   url.Host,
			Url:      getCodeCommitCloneUrl(g.region, reponame),
			Path:     &branchpath,
			Branch:   &branch,
			Sha:      nil,
			PrNumber: nil,
		}, nil
	}
	path := strings.TrimPrefix(url.Path, "/codesuite/codecommit/repositories/")
	parts := strings.Split(path, "/")
	reponame := parts[0]
	staticContext := &StaticGitContext{
		Id:       reponame,
		Name:     reponame,
		Owner:    reponame,
		Source:   url.Host,
		Branch:   &branch,
		Url:      getCodeCommitCloneUrl(g.region, reponame),
		Sha:      nil,
		PrNumber: nil,
		Path:     nil,
	}
	if len(parts) > 2 {
		switch parts[1] {
		case "browse":
			if len(parts) > 4 && parts[3] == "heads" {
				branchpath := fmt.Sprintf("refs/heads/%s", parts[4])
				if len(parts) > 5 && parts[5] == "--" {
					branchpath = fmt.Sprintf("refs/heads/%s/--/%s", parts[4], parts[6])
				}
				staticContext.Branch = &parts[4]
				staticContext.Path = &branchpath
			}
		case "commit":
			sha := parts[2]
			staticContext.Sha = &sha
		case "pull-requests":
			prNumber, err := strconv.ParseUint(parts[2], 10, 32)
			if err == nil {
				prNum := uint32(prNumber)
				staticContext.PrNumber = &prNum
			}
		}
	}
	return staticContext, nil
}

func getCodeCommitCloneUrl(region string, repositoryId string) string {
	return fmt.Sprintf("https://git-codecommit.%s.amazonaws.com/v1/repos/%s", region, repositoryId)
}
