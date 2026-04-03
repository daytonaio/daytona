// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package dto

import specsgen "github.com/daytonaio/runner/pkg/runner/v2/specs/gen"

type RegistryDTO = specsgen.RegistryInfo

func RegistryHasAuth(r *RegistryDTO) bool {
	return r.Username != nil && r.Password != nil && *r.Username != "" && *r.Password != ""
}
