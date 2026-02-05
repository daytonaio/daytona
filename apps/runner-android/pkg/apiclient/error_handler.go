// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package apiclient

import (
	"fmt"
	"io"
	"net/http"

	apiclient "github.com/daytonaio/apiclient"
)

// HandleErrorResponse extracts error details from an API error response
func HandleErrorResponse(err error, resp *http.Response) error {
	if err == nil {
		return nil
	}

	// Try to get more details from the response body
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
		body, readErr := io.ReadAll(resp.Body)
		if readErr == nil && len(body) > 0 {
			return fmt.Errorf("%w: %s", err, string(body))
		}
	}

	// Check if it's a GenericOpenAPIError
	if apiErr, ok := err.(*apiclient.GenericOpenAPIError); ok {
		return fmt.Errorf("%s: %s", apiErr.Error(), string(apiErr.Body()))
	}

	return err
}
