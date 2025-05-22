// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package middlewares

import (
	"context"
	"fmt"
	"strings"

	"github.com/daytonaio/runner-docker/pkg/common"
	"github.com/docker/docker/errdefs"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ErrorInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			var grpcErr error

			switch e := err.(type) {
			case *common.CustomError:
				grpcErr = status.Error(codes.Code(e.StatusCode), e.Message)
			case *common.NotFoundError:
				grpcErr = status.Error(codes.NotFound, e.Error())
			case *common.UnauthorizedError:
				grpcErr = status.Error(codes.Unauthenticated, e.Error())
			case *common.InvalidBodyRequestError:
				grpcErr = status.Error(codes.InvalidArgument, e.Error())
			case *common.ConflictError:
				grpcErr = status.Error(codes.AlreadyExists, e.Error())
			case *common.BadRequestError:
				grpcErr = status.Error(codes.InvalidArgument, e.Error())
			default:
				grpcErr = handlePossibleDockerError(err, info.FullMethod)
			}

			// Log the error
			if status.Code(grpcErr) == codes.Internal {
				log.WithError(err).WithFields(log.Fields{
					"method": info.FullMethod,
				}).Error("Internal Server Error")
			} else {
				log.WithFields(log.Fields{
					"method": info.FullMethod,
					"error":  grpcErr.Error(),
				}).Error("gRPC ERROR")
			}

			return nil, grpcErr
		}
		return resp, nil
	}
}

func handlePossibleDockerError(err error, method string) error {
	if errdefs.IsNotFound(err) {
		return status.Error(codes.NotFound, fmt.Sprintf("resource not found: %s", err.Error()))
	} else if errdefs.IsUnauthorized(err) || strings.Contains(err.Error(), "unauthorized") {
		return status.Error(codes.Unauthenticated, fmt.Sprintf("unauthorized: %s", err.Error()))
	} else if errdefs.IsConflict(err) {
		return status.Error(codes.AlreadyExists, fmt.Sprintf("conflict: %s", err.Error()))
	} else if errdefs.IsInvalidParameter(err) {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("bad request: %s", err.Error()))
	} else if errdefs.IsSystem(err) {
		if strings.Contains(err.Error(), "unable to find user") {
			return status.Error(codes.InvalidArgument, err.Error())
		}
	}

	return status.Error(codes.Internal, err.Error())
}
