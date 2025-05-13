// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package apiclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/daytonaio/daytona/cli/internal"

	log "github.com/sirupsen/logrus"
)

const API_VERSION_HEADER = "X-Daytona-Api-Version"

type ApiErrorResponse struct {
	Error   string `json:"error"`
	Message any    `json:"message,omitempty"`
}

func HandleErrorResponse(res *http.Response, requestErr error) error {
	if res == nil {
		return requestErr
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var errResponse ApiErrorResponse
	err = json.Unmarshal(body, &errResponse)
	if err != nil {
		return err
	}

	checkVersionsMismatch(res)

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

	if res.StatusCode == http.StatusUnauthorized {
		errMessage += " - run 'daytona login' to reauthenticate"
	}

	return errors.New(errMessage)
}

func checkVersionsMismatch(res *http.Response) {
	serverVersion := res.Header.Get(API_VERSION_HEADER)
	if serverVersion == "" {
		return
	}
	if internal.Version != "v0.0.0-dev" && internal.Version != serverVersion {
		log.Warn(fmt.Sprintf("Version mismatch detected. Daytona CLI is on version %s and API is on version %s. To ensure maximum compatibility, please make sure the versions are aligned.", internal.Version, serverVersion))
	}
}
