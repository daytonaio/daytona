// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package models

type SigningMethod string // @name SigningMethod

const (
	SigningMethodSSH SigningMethod = "ssh"
	SigningMethodGPG SigningMethod = "gpg"
)

type GitProviderConfig struct {
	Id            string         `json:"id" validate:"required" gorm:"primaryKey"`
	ProviderId    string         `json:"providerId" validate:"required" gorm:"not null"`
	Username      string         `json:"username" validate:"required" gorm:"not null"`
	BaseApiUrl    *string        `json:"baseApiUrl,omitempty" validate:"optional"`
	Token         string         `json:"token" validate:"required" gorm:"not null"`
	Alias         string         `json:"alias" validate:"required" gorm:"uniqueIndex;not null"`
	SigningKey    *string        `json:"signingKey,omitempty" validate:"optional"`
	SigningMethod *SigningMethod `json:"signingMethod,omitempty" validate:"optional"`
} // @name GitProvider
