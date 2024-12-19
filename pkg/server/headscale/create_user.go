// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package headscale

import (
	v1 "github.com/juanfont/headscale/gen/go/headscale/v1"

	log "github.com/sirupsen/logrus"
)

const HEADSCALE_USERNAME = "daytona"

func (s *HeadscaleServer) CreateUser(username string) error {
	log.Debug("Creating headscale user")

	ctx, client, conn, cancel, err := s.getClient()
	if err != nil {
		return err
	}
	defer cancel()
	defer conn.Close()

	_, err = client.GetUser(ctx, &v1.GetUserRequest{
		Name: username,
	})
	if err == nil {
		log.Debug("User already exists")
		return nil
	}

	_, err = client.CreateUser(ctx, &v1.CreateUserRequest{
		Name: username,
	})

	return err
}
