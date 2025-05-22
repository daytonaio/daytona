// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package middlewares

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func LoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		latency := time.Since(start)
		statusCode := codes.OK
		if err != nil {
			if s, ok := status.FromError(err); ok {
				statusCode = s.Code()
			} else {
				statusCode = codes.Unknown
			}
		}

		log.WithFields(log.Fields{
			"method":     info.FullMethod,
			"status":     statusCode,
			"latency_ms": latency.Milliseconds(),
		}).Info("gRPC request")

		return resp, err
	}
}
