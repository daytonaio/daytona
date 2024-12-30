// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ssh

import (
	"fmt"
)

func (c *Client) WriteFile(content, filePath string) ([]byte, error) {
	session, err := c.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	return session.CombinedOutput(fmt.Sprintf("echo '%s' > %s", content, filePath))
}
