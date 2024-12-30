// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ssh

import (
	"io"
)

func (c *Client) Exec(command string, logWriter io.Writer) error {
	session, err := c.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdout = logWriter
	session.Stderr = logWriter

	return session.Run(command)
}
