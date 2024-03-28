// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/spf13/cobra"
)

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

		encodedUrl := url.QueryEscape(host)
		gitProvider, res, err := apiClient.GitProviderAPI.GetGitProviderForUrl(ctx, encodedUrl).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		if gitProvider == nil {
			fmt.Println("error: git provider not found")
			os.Exit(1)
			return
		}

		fmt.Println("username=" + *gitProvider.Username)
		fmt.Println("password=" + *gitProvider.Token)
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
