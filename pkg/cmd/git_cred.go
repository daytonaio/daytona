// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/spf13/cobra"
)

type GitCredentials struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}

var gitCredCmd = &cobra.Command{
	Use:     "git-cred get",
	Aliases: []string{"rev"},
	Args:    cobra.ExactArgs(1),
	Hidden:  true,
	Run: func(cmd *cobra.Command, args []string) {
		if args[0] != "get" {
			return
		}
		ctx := context.Background()
		result, err := parseFromStdin()
		host := result["host"]
		if err != nil || host == "" {
			fmt.Println("error parsing 'host' from stdin")
			return
		}

		apiClient, err := server.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		var gitCredentials GitCredentials
		serverConfig, res, err := apiClient.ServerAPI.GetConfig(ctx).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		gitProviderId := getGitProviderIdFromHost(host)
		if gitProviderId == "" {
			fmt.Println("error: unable to determine git provider")
			return
		}

		var gitProvider *serverapiclient.GitProvider
		for _, provider := range serverConfig.GitProviders {
			if *provider.Id == gitProviderId {
				gitProvider = &provider
				break
			}
		}

		if gitProvider == nil {
			fmt.Println("error: git provider not found")
			os.Exit(1)
			return
		}

		gitCredentials = GitCredentials{
			Username: *gitProvider.Username,
			Token:    *gitProvider.Token,
		}

		fmt.Println("username=" + gitCredentials.Username)
		fmt.Println("password=" + gitCredentials.Token)
	},
}

func getGitProviderIdFromHost(url string) string {
	if strings.Contains(url, "github.com") {
		return "github"
	} else if strings.Contains(url, "gitlab.com") {
		return "gitlab"
	} else if strings.Contains(url, "bitbucket.org") {
		return "bitbucket"
	} else {
		return ""
	}
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
