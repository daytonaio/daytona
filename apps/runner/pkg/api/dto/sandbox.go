// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package dto

import specsgen "github.com/daytonaio/runner/pkg/runner/v2/specs/gen"

type CreateSandboxDTO struct {
	*specsgen.CreateSandboxPayload
} //	@name	CreateSandboxDTO

type ResizeSandboxDTO struct {
	*specsgen.ResizeSandboxPayload
} //	@name	ResizeSandboxDTO

type UpdateNetworkSettingsDTO struct {
	*specsgen.UpdateNetworkSettingsPayload
} //	@name	UpdateNetworkSettingsDTO

type RecoverSandboxDTO struct {
	*specsgen.RecoverSandboxPayload
} //	@name	RecoverSandboxDTO

type IsRecoverableDTO struct {
	ErrorReason string `json:"errorReason" validate:"required"`
} //	@name	IsRecoverableDTO

type IsRecoverableResponse struct {
	Recoverable bool `json:"recoverable"`
} //	@name	IsRecoverableResponse

type StartSandboxResponse struct {
	DaemonVersion string `json:"daemonVersion"`
} //	@name	StartSandboxResponse

type StopSandboxDTO struct {
	*specsgen.StopSandboxPayload
} //	@name	StopSandboxDTO
