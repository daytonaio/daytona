// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package apiclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/daytonaio/daytona/cli/internal/clierr"
)

type ApiErrorResponse struct {
	Error   string `json:"error"`
	Message any    `json:"message,omitempty"`
}

func HandleErrorResponse(res *http.Response, requestErr error) error {
	if res == nil {
		if requestErr == nil {
			return nil
		}
		return clierr.New(clierr.CategoryNetwork, requestErr.Error())
	}

	defer res.Body.Close()

	// 2xx: error is client-side (e.g. decode failure), not a server error
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return requestErr
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var errResponse ApiErrorResponse
	err = json.Unmarshal(body, &errResponse)
	if err != nil {
		return err
	}

	errMessage := string(errResponse.Error)
	if errMessage == "" {
		// Fall back to raw body if error field is empty
		errMessage = string(body)
	} else {
		if errResponse.Message != nil {
			// Message field could be a string or an array
			switch msg := errResponse.Message.(type) {
			case string:
				errMessage += ": " + msg
			case []any:
				if len(msg) > 0 {
					msgStr := fmt.Sprintf("%v", msg)
					errMessage += ": " + msgStr
				}
			}
		}
	}

	// FromHTTPStatus attaches the 401/403 remediation hints; Error()
	// re-appends them so the rendered text matches the previous output.
	return clierr.FromHTTPStatus(res.StatusCode, errMessage)
}
