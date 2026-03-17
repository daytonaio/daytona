// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package pty

func (cl *wsClient) close() {
	cl.closeOnce.Do(func() {
		close(cl.done)
		_ = cl.conn.Close()
	})
}
