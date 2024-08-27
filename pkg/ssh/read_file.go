// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ssh

import (
	"fmt"
)

func (c *Client) ReadFile(filePath string) ([]byte, error) {
	session, err := c.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	output, err := session.CombinedOutput(fmt.Sprintf("cat %s", filePath))
	if err != nil {
		return nil, err
	}

	return output, nil
}
