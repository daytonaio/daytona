// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;

/**
 * Raised for network-level connection failures (no HTTP response received).
 *
 * <p>Raised when the SDK cannot reach the Daytona API due to network issues
 * such as DNS failure, connection refused, or TLS errors.
 *
 * <pre>{@code
 * try {
 *     daytona.sandbox().create();
 * } catch (DaytonaConnectionException e) {
 *     System.err.println("Cannot reach Daytona API: " + e.getMessage());
 * }
 * }</pre>
 */
public class DaytonaConnectionException extends DaytonaException {
    /**
     * Creates a connection exception.
     *
     * @param message connection failure description
     */
    public DaytonaConnectionException(String message) {
        super(message);
    }

    /**
     * Creates a connection exception with a cause.
     *
     * @param message connection failure description
     * @param cause root cause
     */
    public DaytonaConnectionException(String message, Throwable cause) {
        super(message, cause);
    }
}
