// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/daemon/pkg/common"
)

func classifyA11yError(err error) error {
	msg := err.Error()
	switch {
	case hasA11ySentinel(msg, a11yMsgUnavailable):
		return common.NewA11yUnavailableError(msg)
	case hasA11ySentinel(msg, a11yMsgNodeNotFound),
		hasA11ySentinel(msg, a11yMsgNoAccessibleRoot):
		return common_errors.NewNotFoundError(err)
	case hasA11ySentinel(msg, a11yMsgActionNotSupported),
		hasA11ySentinel(msg, a11yMsgInvalidScope),
		hasA11ySentinel(msg, a11yMsgInvalidRequest):
		return common_errors.NewBadRequestError(err)
	default:
		return common_errors.NewInternalServerError(err)
	}
}
