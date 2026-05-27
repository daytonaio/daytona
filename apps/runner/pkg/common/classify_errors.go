// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"fmt"
	"strings"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/client"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/runner/internal/util"
)

// ClassifyDockerError maps raw Docker / containerd errors to typed errors.
// Returns nil if err isn't a known shape. Used by runnerDefaultErrorHandler.
func ClassifyDockerError(err error) error {
	switch {
	case client.IsErrConnectionFailed(err):
		return NewDockerDaemonUnreachableError(err.Error())
	case errdefs.IsUnauthorized(err) || strings.Contains(err.Error(), "unauthorized"):
		return common_errors.NewUnauthorizedError(fmt.Errorf("unauthorized: %s", err.Error()))
	case errdefs.IsConflict(err):
		return common_errors.NewConflictError(fmt.Errorf("conflict: %s", err.Error()))
	case errdefs.IsInvalidArgument(err):
		return common_errors.NewBadRequestError(fmt.Errorf("bad request: %s", err.Error()))
	case errdefs.IsNotFound(err):
		return common_errors.NewNotFoundError(fmt.Errorf("not found: %s", err.Error()))
	case errdefs.IsInternal(err) && strings.Contains(err.Error(), "unable to find user"):
		return common_errors.NewBadRequestError(fmt.Errorf("%s", util.ExtractErrorPart(err.Error())))
	}
	return nil
}
