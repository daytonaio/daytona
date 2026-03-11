// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"encoding/json"
	"errors"

	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
)

func ConvertOpenAPIError(err error) error {
	if err == nil {
		return nil
	}

	openapiErr := &apiclient.GenericOpenAPIError{}
	if !errors.As(err, &openapiErr) {
		return err
	}

	bodyString := string(openapiErr.Body())

	daytonaErr := &ErrorResponse{}
	if parseErr := json.Unmarshal([]byte(bodyString), daytonaErr); parseErr != nil {
		return err
	}

	return NewCustomError(daytonaErr.StatusCode, daytonaErr.Message, daytonaErr.Code)
}

func IsRetryableOpenAPIError(err error) bool {
	if err == nil {
		return false
	}

	if customErr, ok := err.(*CustomError); ok {
		return customErr.IsRetryable()
	}

	return true
}
