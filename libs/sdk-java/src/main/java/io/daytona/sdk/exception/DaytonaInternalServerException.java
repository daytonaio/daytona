// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * Raised for HTTP 500 — server-side bug or unhandled condition.
 */
public class DaytonaInternalServerException extends DaytonaServerException {
    public static final int STATUS_CODE = 500;

    public DaytonaInternalServerException(String message) {
        super(STATUS_CODE, message);
    }

    public DaytonaInternalServerException(String message, Throwable cause) {
        super(STATUS_CODE, message, cause);
    }

    public DaytonaInternalServerException(String message, String code, String source) {
        super(STATUS_CODE, message, code, source);
    }

    public DaytonaInternalServerException(String message, Throwable cause, String code, String source) {
        super(STATUS_CODE, message, cause, code, source);
    }
}
