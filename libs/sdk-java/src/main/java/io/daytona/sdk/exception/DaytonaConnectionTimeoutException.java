// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;

/**
 * Raised when the transport layer times out connecting to or reading from a
 * Daytona service. Subclass of {@link DaytonaConnectionException} so callers
 * can catch the broader "connection failed" category.
 */
public class DaytonaConnectionTimeoutException extends DaytonaConnectionException {
    public DaytonaConnectionTimeoutException(String message) {
        super(message);
    }

    public DaytonaConnectionTimeoutException(String message, Throwable cause) {
        super(message, cause);
    }

    public DaytonaConnectionTimeoutException(String message, String code, String source) {
        super(message, code, source);
    }

    public DaytonaConnectionTimeoutException(String message, Throwable cause, String code, String source) {
        super(message, cause, code, source);
    }
}
