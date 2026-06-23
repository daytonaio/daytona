// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import "sync"

var x11ClientMu sync.Mutex

func withX11Client(fn func() error) error {
	x11ClientMu.Lock()
	defer x11ClientMu.Unlock()

	return fn()
}
