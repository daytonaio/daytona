// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package git

import (
	"io"

	"github.com/go-git/go-git/v5"
)

type Service struct {
	WorkDir           string
	GitConfigFileName string
	LogWriter         io.Writer
	OpenRepository    *git.Repository
}
