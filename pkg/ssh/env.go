// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ssh

import (
	"io"
	"strings"
)

func (c *Client) GetEnv(logWriter io.Writer) ([]string, error) {
	session, err := c.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	session.Stdout = logWriter
	session.Stderr = logWriter

	out, err := session.CombinedOutput("env")
	if err != nil {
		return nil, err
	}

	output := strings.TrimRight(string(out), "\n")
	return strings.Split(output, "\n"), nil
}
