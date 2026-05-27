// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * A default region must be configured for this operation.
 *
 * <p>Subclass of {@link DaytonaBadRequestException}.
 */
public class DaytonaDefaultRegionRequiredException extends DaytonaBadRequestException {
    public DaytonaDefaultRegionRequiredException(String message) {
        super(message);
    }

    public DaytonaDefaultRegionRequiredException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaDefaultRegionRequiredException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaDefaultRegionRequiredException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
