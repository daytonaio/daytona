// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"net/http"
	"regexp"
	"strconv"
)

func GetHTTPStatusCodeAndMessageFromError(err error) (int, string, error) {

	// pattern to match "status code: <3-digit number>. err: <error message>"
	re := regexp.MustCompile(`status code: (\d{3}) err: (.+)`)

	// first match for the regex pattern in the error message
	match := re.FindStringSubmatch(err.Error())

	if len(match) > 2 {
		// matched string to an integer (status code)
		statusCode, convErr := strconv.Atoi(match[1])
		if convErr == nil {
			errorMessage := match[2]
			return statusCode, errorMessage, nil
		}
	}

	return http.StatusInternalServerError, "", err
}
