// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package middlewares

import (
	"context"
	"os"
	"strings"

	"github.com/daytonaio/runner-docker/internal/constants"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	HealthCheckMethod = "/runner.Runner/HealthCheck"
)

func AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip authentication for health check
		if info.FullMethod == HealthCheckMethod {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
		}

		authHeader := md.Get(constants.AUTHORIZATION_HEADER)
		if len(authHeader) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
		}

		authParts := strings.Split(authHeader[0], " ")
		if len(authParts) != 2 || authParts[0] != constants.BEARER_AUTH_HEADER {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization token format")
		}

		token := authParts[1]
		expectedToken := os.Getenv("TOKEN")

		if token != expectedToken {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization token")
		}

		return handler(ctx, req)
	}
}
