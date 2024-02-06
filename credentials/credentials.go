// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package credentials

import "errors"

type Credential struct {
	Username string
	Password string
}

type CredentialProvider interface {
	GetAccessToken() (*Credential, error)
}

type CredentialsClient struct{}

func (c CredentialsClient) GetAccessToken() (*Credential, error) {
	return nil, errors.New("not implemented")
}
