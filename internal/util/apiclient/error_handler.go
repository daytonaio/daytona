// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apiclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/daytonaio/daytona/internal"
	"github.com/daytonaio/daytona/pkg/api/middlewares"
	log "github.com/sirupsen/logrus"
)

type ApiErrorResponse struct {
	Error string `json:"error"`
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
		return errors.New(string(body))
	}

	if !IsHealthCheckFailed(errors.New(errResponse.Error)) {
		checkVersionsMismatch(res)
	}

	return errors.New(errResponse.Error)
}

func checkVersionsMismatch(res *http.Response) {
	serverVersion := res.Header.Get(middlewares.SERVER_VERSION_HEADER)
	if internal.Version != serverVersion {
		log.Warn(fmt.Sprintf("Version mismatch detected. CLI is on version %s, Daytona Server is on version %s. To ensure maximum compatibility, please make sure the versions are aligned.", internal.Version, serverVersion))
	}
}
