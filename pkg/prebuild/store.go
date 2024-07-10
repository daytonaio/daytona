// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package prebuild

import "errors"

type PrebuildFilter struct {
	ProjectConfigName string
}

type ConfigStore interface {
	Find(key string) (*Prebuild, error)
	Save(p *Prebuild) error
	List(filter *PrebuildFilter) ([]*Prebuild, error)
	Delete(p *Prebuild) error
}

var (
	ErrPrebuildNotFound = errors.New("prebuild config not found")
)

func IsPrebuildNotFound(err error) bool {
	return err.Error() == ErrPrebuildNotFound.Error()
}
