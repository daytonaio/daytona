// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikey

type Store interface {
	List() ([]*ApiKey, error)
	Find(key string) (*ApiKey, error)
	FindByName(name string) (*ApiKey, error)
	Save(apiKey *ApiKey) error
	Delete(apiKey *ApiKey) error
}
