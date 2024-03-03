// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apiclient

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
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
		return err
	}

	return errors.New(errResponse.Error)
}
