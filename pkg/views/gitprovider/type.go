// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

type GitProviderView struct {
	Id                 string
	Name               string
	Username           string
	BaseApiUrl         string
	Token              string
	TokenScopeIdentity string
	TokenScope         string
	TokenScopeType     string
}
