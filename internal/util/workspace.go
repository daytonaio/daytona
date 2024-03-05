// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
)

func WorkspaceMode() bool {
	_, wsMode := os.LookupEnv("DAYTONA_WS_DIR")
	return wsMode
}

func GetFirstWorkspaceProjectName(workspaceId string, projectName string, profile *config.Profile) (string, error) {
	ctx := context.Background()

	apiClient, err := server.GetApiClient(profile)
	if err != nil {
		return "", err
	}

	wsInfo, res, err := apiClient.WorkspaceAPI.GetWorkspace(ctx, workspaceId).Execute()
	if err != nil {
		return "", apiclient.HandleErrorResponse(res, err)
	}

	if projectName == "" {
		if len(wsInfo.Projects) == 0 {
			return "", errors.New("no projects found in workspace")
		}

		return *wsInfo.Projects[0].Name, nil
	}

	for _, project := range wsInfo.Projects {
		if *project.Name == projectName {
			return *project.Name, nil
		}
	}

	return "", errors.New("project not found in workspace")
}

func GetValidatedWorkspaceName(input string) (string, error) {
	// input = strings.ToLower(input)

	input = strings.ReplaceAll(input, " ", "-")

	// Regular expression that catches letters, numbers, and dashes
	pattern := "^[a-zA-Z0-9-]+$"

	matched, err := regexp.MatchString(pattern, input)
	if err != nil {
		return "", err
	}

	if !matched {
		return "", fmt.Errorf("only letters, numbers, and dashes are allowed")
	}

	return input, nil
}

func GetValidatedUrl(input string) (string, error) {
	// Check if the input starts with a scheme (e.g., http:// or https://)
	if !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") {
		return "", fmt.Errorf("input is missing http:// or https://")
	}

	// Try to parse the input as a URL
	parsedURL, err := url.Parse(input)
	if err != nil {
		return "", fmt.Errorf("input is not a valid URL")
	}

	// Validate the URL's host (domain) has a proper extension
	host := parsedURL.Host
	if !isValidTLD(host) {
		return "", fmt.Errorf("the URL does not have a valid TLD")
	}

	// If parsing was successful, return the fixed URL
	return parsedURL.String(), nil
}

func isValidTLD(host string) bool {
	// Regular expression to match common domain extensions like .com, .org, etc.
	extensionPattern := `\.([a-zA-Z]{2,6})$`
	regex := regexp.MustCompile(extensionPattern)

	// Check if the host (domain) matches the extension pattern
	return regex.MatchString(host)
}
