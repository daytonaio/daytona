// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package models

type RunnerServiceInfo struct {
	ServiceName string
	Healthy     bool
	Err         error
}
