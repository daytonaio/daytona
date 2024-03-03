// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package headscale

import (
	"fmt"
	"time"

	v1 "github.com/juanfont/headscale/gen/go/headscale/v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	log "github.com/sirupsen/logrus"
)

func CreateAuthKey() (string, error) {
	log.Debug("Creating headscale auth key")

	request := &v1.CreatePreAuthKeyRequest{
		Reusable: false,
		User:     "daytona",
	}
	request.Expiration = timestamppb.New(time.Now().Add(100000 * time.Hour))
	request.Ephemeral = true

	ctx, client, conn, cancel, err := getClient()
	if err != nil {
		return "", fmt.Errorf("failed to get client: %w", err)
	}
	defer cancel()
	defer conn.Close()

	response, err := client.CreatePreAuthKey(ctx, request)
	if err != nil {
		return "", fmt.Errorf("failed to create ApiKey: %w", err)
	}

	log.Debug("Headscale auth key created")

	return response.PreAuthKey.Key, nil
}
