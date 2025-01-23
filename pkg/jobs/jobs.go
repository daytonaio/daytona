// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package jobs

import (
	"context"
)

type IJob interface {
	Execute(ctx context.Context) error
}
