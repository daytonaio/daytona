// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apiclient

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/daytonaio/daytona/internal"
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

	version := res.Header.Get("X-Server-Version")
	if internal.Version != version {
		log.Warn("Version mismatch! Server and client version are not the same. Please, update your Daytona versions.")
	}

	var errResponse ApiErrorResponse

	err = json.Unmarshal(body, &errResponse)
	if err != nil {
		return errors.New(string(body))
	}

	return errors.New(errResponse.Error)
}
