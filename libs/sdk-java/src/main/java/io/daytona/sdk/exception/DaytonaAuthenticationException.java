// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;

/**
 * Raised when API credentials are missing or invalid (HTTP 401).
 *
 * <pre>{@code
 * try {
 *     daytona.sandbox().create();
 * } catch (DaytonaAuthenticationException e) {
 *     System.err.println("Invalid or missing API key");
 * }
 * }</pre>
 */
public class DaytonaAuthenticationException extends DaytonaException {
    /** HTTP status code carried by every instance of this class. */
    public static final int STATUS_CODE = 401;

    /**
     * Creates an authentication exception.
     *
     * @param message error description from the API
     */
    public DaytonaAuthenticationException(String message) {
        super(STATUS_CODE, message);
    }

    /**
     * @param message error description from the API
     * @param cause root cause
     */
    public DaytonaAuthenticationException(String message, Throwable cause) {
        super(STATUS_CODE, message, cause);
    }

    public DaytonaAuthenticationException(String message, String code, String source) {
        super(STATUS_CODE, message, code, source);
    }

    public DaytonaAuthenticationException(String message, Throwable cause, String code, String source) {
        super(STATUS_CODE, message, cause, code, source);
    }
}
