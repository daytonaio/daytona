// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package dto

type VolumeDTO struct {
	Id        string `json:"id" validate:"required"`
	MountPath string `json:"mountPath" validate:"required"`
} //	@name	VolumeDTO



