// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package io.daytona.sdk.exception;
/**
 * Raised for HTTP 410 — the target resource is permanently gone.
 */
public class DaytonaGoneException extends DaytonaException {
    public static final int STATUS_CODE = 410;

    public DaytonaGoneException(String message) {
        super(STATUS_CODE, message);
    }

    public DaytonaGoneException(String message, Throwable cause) {
        super(STATUS_CODE, message, cause);
    }

    public DaytonaGoneException(String message, String code, String source) {
        super(STATUS_CODE, message, code, source);
    }

    public DaytonaGoneException(String message, Throwable cause, String code, String source) {
        super(STATUS_CODE, message, cause, code, source);
    }
}
