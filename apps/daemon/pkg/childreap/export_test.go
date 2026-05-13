// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package childreap

import "time"

// overrideRecoveryTimeoutForTest swaps in a shorter timeout for tests that
// need to exercise the timeout path without burning wall time. Returns the
// previous value so the caller can restore it.
func overrideRecoveryTimeoutForTest(d time.Duration) time.Duration {
	prev := recoveryTimeout
	recoveryTimeout = d
	return prev
}
