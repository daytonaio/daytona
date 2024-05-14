// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profiledata

type Store interface {
	Get() (*ProfileData, error)
	Save(profileData *ProfileData) error
	Delete() error
}
