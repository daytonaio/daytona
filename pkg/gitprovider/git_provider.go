// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

const personalNamespaceId = "<PERSONAL>"

type GitProvider interface {
	GetNamespaces() ([]*GitNamespace, error)
	GetRepositories(namespace string) ([]*GitRepository, error)
	GetUser() (*GitUser, error)
	GetRepoBranches(repositoryId string, namespaceId string) ([]*GitBranch, error)
	GetRepoPRs(repositoryId string, namespaceId string) ([]*GitPullRequest, error)
	// ParseGitUrl(string) (*GitRepository, error)
}
