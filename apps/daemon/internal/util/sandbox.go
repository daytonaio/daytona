// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package util

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

func GetValidatedName(input string) (string, error) {
	input = strings.ReplaceAll(input, " ", "-")

	// Regular expression that catches letters, numbers, and dashes
	pattern := "^[a-zA-Z0-9-]+$"

	matched, err := regexp.MatchString(pattern, input)
	if err != nil {
		return "", err
	}

	if !matched {
		return "", errors.New("only letters, numbers, and dashes are allowed")
	}

	return input, nil
}

func GetValidatedUrl(input string) (string, error) {
	// Check if the input starts with a scheme (e.g., http:// or https://)
	if !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") {
		return "", errors.New("input is missing http:// or https://")
	}

	// Try to parse the input as a URL
	parsedURL, err := url.Parse(input)
	if err != nil {
		return "", errors.New("input is not a valid URL")
	}

	// If parsing was successful, return the fixed URL
	return parsedURL.String(), nil
}

func GetRepositorySlugFromUrl(url string, specifyGitProviders bool) string {
	if url == "" {
		return "/"
	}
	url = strings.TrimSuffix(url, "/")

	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return ""
	}

	if specifyGitProviders {
		return parts[len(parts)-3] + "/" + parts[len(parts)-2] + "/" + parts[len(parts)-1]
	}

	return parts[len(parts)-2] + "/" + parts[len(parts)-1]
}

func CleanUpRepositoryUrl(url string) string {
	url = strings.ToLower(url)
	return strings.TrimSuffix(url, "/")
}
