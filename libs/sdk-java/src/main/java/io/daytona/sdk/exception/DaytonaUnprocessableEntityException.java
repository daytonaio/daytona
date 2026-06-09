// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;

/**
 * Raised for HTTP 422 — the request is well-formed but semantically invalid
 * (e.g. unsupported resource class, invalid configuration values).
 *
 * <pre>{@code
 * try {
 *     daytona.sandbox().create(params);
 * } catch (DaytonaUnprocessableEntityException e) {
 *     System.err.println("Unprocessable entity: " + e.getMessage());
 * }
 * }</pre>
 */
public class DaytonaUnprocessableEntityException extends DaytonaException {
    /** HTTP status code carried by every instance of this class. */
    public static final int STATUS_CODE = 422;

    public DaytonaUnprocessableEntityException(String message) {
        super(STATUS_CODE, message);
    }

    public DaytonaUnprocessableEntityException(String message, Throwable cause) {
        super(STATUS_CODE, message, cause);
    }

    public DaytonaUnprocessableEntityException(String message, String code, String source) {
        super(STATUS_CODE, message, code, source);
    }

    public DaytonaUnprocessableEntityException(String message, Throwable cause, String code, String source) {
        super(STATUS_CODE, message, cause, code, source);
    }
}
