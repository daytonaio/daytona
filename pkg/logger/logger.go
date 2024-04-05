// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package logger

import (
	"io"
)

type Logger interface {
	io.WriteCloser
	Cleanup() error
}
