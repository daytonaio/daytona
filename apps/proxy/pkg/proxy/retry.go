// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import "time"

const (
	proxyMaxRetries = 3
	proxyBaseDelay  = 150 * time.Millisecond
	proxyMaxDelay   = 1 * time.Second
)
