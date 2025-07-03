// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"fmt"
	"strings"

	"github.com/docker/docker/errdefs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func MapDockerError(err error) error {
	if errdefs.IsNotFound(err) {
		return status.Error(codes.NotFound, fmt.Sprintf("resource not found: %s", err.Error()))
	}

	if errdefs.IsUnauthorized(err) || strings.Contains(err.Error(), "unauthorized") {
		return status.Error(codes.Unauthenticated, fmt.Sprintf("unauthorized: %s", err.Error()))
	}

	if errdefs.IsConflict(err) {
		return status.Error(codes.AlreadyExists, fmt.Sprintf("conflict: %s", err.Error()))
	}

	if errdefs.IsInvalidParameter(err) {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("bad request: %s", err.Error()))
	}

	if errdefs.IsSystem(err) {
		if strings.Contains(err.Error(), "unable to find user") {
			return status.Error(codes.InvalidArgument, err.Error())
		}
	}

	return status.Error(codes.Internal, err.Error())
}
