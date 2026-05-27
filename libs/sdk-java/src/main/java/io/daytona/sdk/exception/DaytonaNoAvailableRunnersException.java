// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * No runner is available to satisfy the request.
 *
 * <p>Subclass of {@link DaytonaBadRequestException}.
 */
public class DaytonaNoAvailableRunnersException extends DaytonaBadRequestException {
    public DaytonaNoAvailableRunnersException(String message) {
        super(message);
    }

    public DaytonaNoAvailableRunnersException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaNoAvailableRunnersException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaNoAvailableRunnersException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
