// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package interceptors

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

// GetDefaultInterceptors returns a slice of default interceptors for the gRPC server
func GetDefaultInterceptors() []grpc.UnaryServerInterceptor {
	logger := zerolog.New(zerolog.NewConsoleWriter())

	return []grpc.UnaryServerInterceptor{
		auth.UnaryServerInterceptor(authFn),
		logging.UnaryServerInterceptor(interceptorLogger(&logger)),
	}
}

// ChainUnaryServer creates a single interceptor out of a chain of many interceptors
func ChainUnaryServer(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		chain := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			chain = buildUnaryServerInterceptor(interceptors[i], chain)
		}
		return chain(ctx, req)
	}
}

func buildUnaryServerInterceptor(interceptor grpc.UnaryServerInterceptor, next grpc.UnaryHandler) grpc.UnaryHandler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return interceptor(ctx, req, nil, next)
	}
}
