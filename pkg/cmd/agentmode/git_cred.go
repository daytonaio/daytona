// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agentmode

import (
	"bufio"
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/spf13/cobra"
)

var gitCredCmd = &cobra.Command{
	Use:     "git-cred get",
	Aliases: []string{"rev"},
	Args:    cobra.ExactArgs(1),
	Hidden:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if args[0] != "get" {
			return nil
		}
		ctx := context.Background()
		result, err := parseFromStdin()
		host := result["host"]
		if err != nil || host == "" {
			fmt.Println("error parsing 'host' from stdin")
			return nil
		}

		apiClient, err := apiclient.GetApiClient(nil)
		if err != nil {
			return err
		}

		workspace, res, err := apiClient.WorkspaceAPI.GetWorkspace(ctx, workspaceId).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		if workspace.GitProviderConfigId != nil {
			gitProvider, _, _ := apiClient.GitProviderAPI.GetGitProvider(ctx, *workspace.GitProviderConfigId).Execute()
			if gitProvider != nil {
				fmt.Println("username=" + gitProvider.Username)
				fmt.Println("password=" + gitProvider.Token)
				return nil
			}

		}

		encodedUrl := url.QueryEscape(host)
		gitProviders, _, _ := apiClient.GitProviderAPI.ListGitProvidersForUrl(ctx, encodedUrl).Execute()
		if len(gitProviders) == 0 {
			fmt.Println("error: git provider not found")
			os.Exit(1)
		}

		fmt.Println("username=" + gitProviders[0].Username)
		fmt.Println("password=" + gitProviders[0].Token)
		return nil
	},
}

func parseFromStdin() (map[string]string, error) {
	result := make(map[string]string)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) > 0 {
			tuple := strings.Split(line, "=")
			if len(tuple) == 2 {
				result[tuple[0]] = strings.TrimSpace(tuple[1])
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
