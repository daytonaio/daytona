// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"slices"
	"strings"

	"github.com/daytonaio/daytona/internal/jetbrains"
	"github.com/daytonaio/daytona/pkg/os"
)

func GetBinaryUrls() map[os.OperatingSystem]string {
	return map[os.OperatingSystem]string{
		(os.Darwin_64_86):  "https://download.daytona.io/daytona/latest/daytona-darwin-amd64",
		(os.Darwin_arm64):  "https://download.daytona.io/daytona/latest/daytona-darwin-arm64",
		(os.Linux_64_86):   "https://download.daytona.io/daytona/latest/daytona-linux-amd64",
		(os.Linux_arm64):   "https://download.daytona.io/daytona/latest/daytona-linux-arm64",
		(os.Windows_64_86): "https://download.daytona.io/daytona/latest/daytona-windows-amd64.exe",
		(os.Windows_arm64): "https://download.daytona.io/daytona/latest/daytona-windows-arm64.exe",
	}
}

func GetIdeList() []Ide {
	ides := []Ide{
		{"vscode", "VS Code"},
		{"browser", "VS Code - Browser"},
		{"cursor", "Cursor"},
		{"ssh", "Terminal SSH"},
		{"jupyter", "Jupyter"},
		{"fleet", "Fleet"},
	}

	sortedJbIdes := []Ide{}
	for id, ide := range jetbrains.GetIdes() {
		sortedJbIdes = append(sortedJbIdes, Ide{string(id), ide.Name})
	}
	slices.SortFunc(sortedJbIdes, func(i, j Ide) int {
		return strings.Compare(i.Name, j.Name)
	})
	ides = append(ides, sortedJbIdes...)

	return ides
}

func GetSupportedGitProviders() []GitProvider {
	return []GitProvider{
		{"github", "GitHub"},
		{"github-enterprise-server", "GitHub Enterprise Server"},
		{"gitlab", "GitLab"},
		{"gitlab-self-managed", "GitLab Self-managed"},
		{"bitbucket", "Bitbucket"},
		{"bitbucket-server", "Bitbucket Server"},
		{"codeberg", "Codeberg"},
		{"gitea", "Gitea"},
		{"gitness", "Gitness"},
		{"azure-devops", "Azure DevOps"},
		{"aws-codecommit", "AWS CodeCommit"},
	}
}

func GetDocsLinkFromGitProvider(providerId string) string {
	switch providerId {
	case "github":
		fallthrough
	case "github-enterprise-server":
		return "https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens#creating-a-personal-access-token-classic"
	case "gitlab":
		fallthrough
	case "gitlab-self-managed":
		return "https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html#create-a-personal-access-token"
	case "bitbucket":
		return "https://support.atlassian.com/bitbucket-cloud/docs/create-an-app-password"
	case "bitbucket-server":
		return "https://confluence.atlassian.com/bitbucketserver/http-access-tokens-939515499.html"
	case "codeberg":
		return "https://docs.codeberg.org/advanced/access-token/"
	case "gitea":
		return "https://docs.gitea.com/1.21/development/api-usage#generating-and-listing-api-tokens"
	case "gitness":
		return "https://docs.gitness.com/administration/user-management#generate-user-token"
	case "azure-devops":
		return "https://learn.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate?view=azure-devops&tabs=Windows#create-a-pat"
	case "aws-codecommit":
		return "https://docs.aws.amazon.com/codecommit/latest/userguide/setting-up-gc.html and to configure AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY & AWS_DEFAULT_REGION read https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html"
	default:
		return ""
	}
}

func GetRequiredScopesFromGitProviderId(providerId string) string {
	switch providerId {
	case "github":
		fallthrough
	case "github-enterprise-server":
		return "repo,read:user,user:email"
	case "gitlab":
		fallthrough
	case "gitlab-self-managed":
		return "api,read_user,write_repository"
	case "bitbucket":
		return "account:read,repositories:write,pullrequests:read"
	case "bitbucket-server":
		return "PROJECT_READ,REPOSITORY_WRITE"
	case "codeberg":
		fallthrough
	case "gitea":
		return "read:organization,write:repository,read:user"
	case "gitness":
		return "/"
	case "azure-devops":
		return "Code (Status, Read & Write); User Profile (Read); Project and Team (Read)"
	case "aws-codecommit":
		return "/"
	default:
		return ""
	}
}

func GetPrebuildScopesFromGitProviderId(providerId string) string {
	switch providerId {
	case "github":
		fallthrough
	case "github-enterprise-server":
		return "admin:repo_hook"
	case "bitbucket":
		return "webhooks"
	case "azure-devops":
		return "Work (Read, Write & Manage); Build (Read & Execute)"
	default:
		return ""
	}
}

func GetWebhookEventHeaderKeyFromGitProvider(providerId string) string {
	switch providerId {
	case "github":
		return "X-GitHub-Event"
	case "gitlab":
		fallthrough
	case "gitlab-self-managed":
		return "X-Gitlab-Event"
	case "bitbucket":
		return "X-Event-Key"
	case "gitea":
		return "X-Gitea-Event"
	case "azure-devops":
		return "X-AzureDevops-Event"
	case "gitness":
		return "X-Gitness-Trigger"
	default:
		return ""
	}
}