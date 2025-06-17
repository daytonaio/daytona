// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package grpc

import (
	"context"
	"os"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func authFn(ctx context.Context) (context.Context, error) {
	token, err := auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return nil, err
	}
	if token != os.Getenv("TOKEN") {
		return nil, status.Error(codes.Unauthenticated, "invalid auth token")
	}
	return ctx, nil
}
