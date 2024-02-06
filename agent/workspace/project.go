// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"github.com/reactivex/rxgo/v2"
)

type Repository struct {
	Url      string   `json:"url"`
	Branch   *string  `default:"main" json:"branch,omitempty"`
	SHA      *string  `json:"sha,omitempty"`
	Owner    *string  `json:"owner,omitempty"`
	PrNumber *float32 `json:"prNumber,omitempty"`
	Source   *string  `json:"source,omitempty"`
	Path     *string  `json:"path,omitempty"`
}

type Project struct {
	Name       string     `json:"name"`
	Repository Repository `json:"repository"`

	Workspace *Workspace     `json:"-"`
	Events    chan rxgo.Item `json:"-"`
}

type ProjectInfo struct {
	Name                string      `json:"name"`
	Created             string      `json:"created"`
	Started             string      `json:"started"`
	Finished            string      `json:"finished"`
	IsRunning           bool        `json:"isRunning"`
	ProvisionerMetadata interface{} `json:"provisionerMetadata"`
}
