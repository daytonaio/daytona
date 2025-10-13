// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

type Config struct {
	Endpoint       string
	Headers        map[string]string
	ServiceName    string
	ServiceVersion string
}
