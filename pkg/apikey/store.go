// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikey

import "errors"

type Store interface {
	List() ([]*ApiKey, error)
	Find(key string) (*ApiKey, error)
	FindByName(name string) (*ApiKey, error)
	Save(apiKey *ApiKey) error
	Delete(apiKey *ApiKey) error
}

var (
	ErrApiKeyNotFound = errors.New("api key not found")
)

func IsApiKeyNotFound(err error) bool {
	return err.Error() == ErrApiKeyNotFound.Error()
}
