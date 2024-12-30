// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ssh

import (
	"errors"
	"regexp"
)

func (c *Client) GetUserUidGid() (string, string, error) {
	session, err := c.NewSession()
	if err != nil {
		return "", "", err
	}
	defer session.Close()

	idResp, err := session.Output("id")
	if err != nil {
		return "", "", err
	}

	re := regexp.MustCompile(`uid=(\d+).*gid=(\d+)`)
	matches := re.FindStringSubmatch(string(idResp))
	if len(matches) < 3 {
		return "", "", errors.New("could not parse uid and gid from id command")
	}

	return matches[1], matches[2], nil
}
