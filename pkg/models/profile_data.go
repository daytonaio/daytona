// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package models

type ProfileData struct {
	Id      string            `json:"id" validate:"required" gorm:"primaryKey"`
	EnvVars map[string]string `json:"envVars" validate:"required" gorm:"serializer:json"`
} // @name ProfileData
