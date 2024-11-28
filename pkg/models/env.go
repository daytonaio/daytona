// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package models

type EnvironmentVariable struct {
	Key   string `json:"key" validate:"required" gorm:"primaryKey"`
	Value string `json:"value" validate:"required"`
} // @name EnvironmentVariable
